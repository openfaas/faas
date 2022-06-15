// Copyright 2021-2022 The NATS Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package nats

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nuid"
)

// ObjectStoreManager creates, loads and deletes Object Stores
//
// Notice: Experimental Preview
//
// This functionality is EXPERIMENTAL and may be changed in later releases.
type ObjectStoreManager interface {
	// ObjectStore will lookup and bind to an existing object store instance.
	ObjectStore(bucket string) (ObjectStore, error)
	// CreateObjectStore will create an object store.
	CreateObjectStore(cfg *ObjectStoreConfig) (ObjectStore, error)
	// DeleteObjectStore will delete the underlying stream for the named object.
	DeleteObjectStore(bucket string) error
}

// ObjectStore is a blob store capable of storing large objects efficiently in
// JetStream streams
//
// Notice: Experimental Preview
//
// This functionality is EXPERIMENTAL and may be changed in later releases.
type ObjectStore interface {
	// Put will place the contents from the reader into a new object.
	Put(obj *ObjectMeta, reader io.Reader, opts ...ObjectOpt) (*ObjectInfo, error)
	// Get will pull the named object from the object store.
	Get(name string, opts ...ObjectOpt) (ObjectResult, error)

	// PutBytes is convenience function to put a byte slice into this object store.
	PutBytes(name string, data []byte, opts ...ObjectOpt) (*ObjectInfo, error)
	// GetBytes is a convenience function to pull an object from this object store and return it as a byte slice.
	GetBytes(name string, opts ...ObjectOpt) ([]byte, error)

	// PutBytes is convenience function to put a string into this object store.
	PutString(name string, data string, opts ...ObjectOpt) (*ObjectInfo, error)
	// GetString is a convenience function to pull an object from this object store and return it as a string.
	GetString(name string, opts ...ObjectOpt) (string, error)

	// PutFile is convenience function to put a file into this object store.
	PutFile(file string, opts ...ObjectOpt) (*ObjectInfo, error)
	// GetFile is a convenience function to pull an object from this object store and place it in a file.
	GetFile(name, file string, opts ...ObjectOpt) error

	// GetInfo will retrieve the current information for the object.
	GetInfo(name string) (*ObjectInfo, error)
	// UpdateMeta will update the meta data for the object.
	UpdateMeta(name string, meta *ObjectMeta) error

	// Delete will delete the named object.
	Delete(name string) error

	// AddLink will add a link to another object into this object store.
	AddLink(name string, obj *ObjectInfo) (*ObjectInfo, error)

	// AddBucketLink will add a link to another object store.
	AddBucketLink(name string, bucket ObjectStore) (*ObjectInfo, error)

	// Seal will seal the object store, no further modifications will be allowed.
	Seal() error

	// Watch for changes in the underlying store and receive meta information updates.
	Watch(opts ...WatchOpt) (ObjectWatcher, error)

	// List will list all the objects in this store.
	List(opts ...WatchOpt) ([]*ObjectInfo, error)

	// Status retrieves run-time status about the backing store of the bucket.
	Status() (ObjectStoreStatus, error)
}

type ObjectOpt interface {
	configureObject(opts *objOpts) error
}

type objOpts struct {
	ctx context.Context
}

// For nats.Context() support.
func (ctx ContextOpt) configureObject(opts *objOpts) error {
	opts.ctx = ctx
	return nil
}

// ObjectWatcher is what is returned when doing a watch.
type ObjectWatcher interface {
	// Updates returns a channel to read any updates to entries.
	Updates() <-chan *ObjectInfo
	// Stop will stop this watcher.
	Stop() error
}

var (
	ErrObjectConfigRequired = errors.New("nats: object-store config required")
	ErrBadObjectMeta        = errors.New("nats: object-store meta information invalid")
	ErrObjectNotFound       = errors.New("nats: object not found")
	ErrInvalidStoreName     = errors.New("nats: invalid object-store name")
	ErrInvalidObjectName    = errors.New("nats: invalid object name")
	ErrDigestMismatch       = errors.New("nats: received a corrupt object, digests do not match")
	ErrNoObjectsFound       = errors.New("nats: no objects found")
)

// ObjectStoreConfig is the config for the object store.
type ObjectStoreConfig struct {
	Bucket      string
	Description string
	TTL         time.Duration
	MaxBytes    int64
	Storage     StorageType
	Replicas    int
	Placement   *Placement
}

type ObjectStoreStatus interface {
	// Bucket is the name of the bucket
	Bucket() string
	// Description is the description supplied when creating the bucket
	Description() string
	// TTL indicates how long objects are kept in the bucket
	TTL() time.Duration
	// Storage indicates the underlying JetStream storage technology used to store data
	Storage() StorageType
	// Replicas indicates how many storage replicas are kept for the data in the bucket
	Replicas() int
	// Sealed indicates the stream is sealed and cannot be modified in any way
	Sealed() bool
	// Size is the combined size of all data in the bucket including metadata, in bytes
	Size() uint64
	// BackingStore provides details about the underlying storage
	BackingStore() string
}

// ObjectMetaOptions
type ObjectMetaOptions struct {
	Link      *ObjectLink `json:"link,omitempty"`
	ChunkSize uint32      `json:"max_chunk_size,omitempty"`
}

// ObjectMeta is high level information about an object.
type ObjectMeta struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Headers     Header `json:"headers,omitempty"`

	// Optional options.
	Opts *ObjectMetaOptions `json:"options,omitempty"`
}

// ObjectInfo is meta plus instance information.
type ObjectInfo struct {
	ObjectMeta
	Bucket  string    `json:"bucket"`
	NUID    string    `json:"nuid"`
	Size    uint64    `json:"size"`
	ModTime time.Time `json:"mtime"`
	Chunks  uint32    `json:"chunks"`
	Digest  string    `json:"digest,omitempty"`
	Deleted bool      `json:"deleted,omitempty"`
}

// ObjectLink is used to embed links to other buckets and objects.
type ObjectLink struct {
	// Bucket is the name of the other object store.
	Bucket string `json:"bucket"`
	// Name can be used to link to a single object.
	// If empty means this is a link to the whole store, like a directory.
	Name string `json:"name,omitempty"`
}

// ObjectResult will return the underlying stream info and also be an io.ReadCloser.
type ObjectResult interface {
	io.ReadCloser
	Info() (*ObjectInfo, error)
	Error() error
}

const (
	objNameTmpl         = "OBJ_%s"
	objSubjectsPre      = "$O."
	objAllChunksPreTmpl = "$O.%s.C.>"
	objAllMetaPreTmpl   = "$O.%s.M.>"
	objChunksPreTmpl    = "$O.%s.C.%s"
	objMetaPreTmpl      = "$O.%s.M.%s"
	objNoPending        = "0"
	objDefaultChunkSize = uint32(128 * 1024) // 128k
	objDigestType       = "sha-256="
	objDigestTmpl       = objDigestType + "%s"
)

type obs struct {
	name   string
	stream string
	js     *js
}

// CreateObjectStore will create an object store.
func (js *js) CreateObjectStore(cfg *ObjectStoreConfig) (ObjectStore, error) {
	if !js.nc.serverMinVersion(2, 6, 2) {
		return nil, errors.New("nats: object-store requires at least server version 2.6.2")
	}
	if cfg == nil {
		return nil, ErrObjectConfigRequired
	}
	if !validBucketRe.MatchString(cfg.Bucket) {
		return nil, ErrInvalidStoreName
	}

	name := cfg.Bucket
	chunks := fmt.Sprintf(objAllChunksPreTmpl, name)
	meta := fmt.Sprintf(objAllMetaPreTmpl, name)

	scfg := &StreamConfig{
		Name:        fmt.Sprintf(objNameTmpl, name),
		Description: cfg.Description,
		Subjects:    []string{chunks, meta},
		MaxAge:      cfg.TTL,
		MaxBytes:    cfg.MaxBytes,
		Storage:     cfg.Storage,
		Replicas:    cfg.Replicas,
		Placement:   cfg.Placement,
		Discard:     DiscardNew,
		AllowRollup: true,
	}

	// Create our stream.
	_, err := js.AddStream(scfg)
	if err != nil {
		return nil, err
	}

	return &obs{name: name, stream: scfg.Name, js: js}, nil
}

// ObjectStore will lookup and bind to an existing object store instance.
func (js *js) ObjectStore(bucket string) (ObjectStore, error) {
	if !validBucketRe.MatchString(bucket) {
		return nil, ErrInvalidStoreName
	}
	if !js.nc.serverMinVersion(2, 6, 2) {
		return nil, errors.New("nats: key-value requires at least server version 2.6.2")
	}

	stream := fmt.Sprintf(objNameTmpl, bucket)
	si, err := js.StreamInfo(stream)
	if err != nil {
		return nil, err
	}
	return &obs{name: bucket, stream: si.Config.Name, js: js}, nil
}

// DeleteObjectStore will delete the underlying stream for the named object.
func (js *js) DeleteObjectStore(bucket string) error {
	stream := fmt.Sprintf(objNameTmpl, bucket)
	return js.DeleteStream(stream)
}

func sanitizeName(name string) string {
	stream := strings.ReplaceAll(name, ".", "_")
	return strings.ReplaceAll(stream, " ", "_")
}

// Put will place the contents from the reader into this object-store.
func (obs *obs) Put(meta *ObjectMeta, r io.Reader, opts ...ObjectOpt) (*ObjectInfo, error) {
	if meta == nil {
		return nil, ErrBadObjectMeta
	}

	obj := sanitizeName(meta.Name)
	if !keyValid(obj) {
		return nil, ErrInvalidObjectName
	}

	var o objOpts
	for _, opt := range opts {
		if opt != nil {
			if err := opt.configureObject(&o); err != nil {
				return nil, err
			}
		}
	}
	ctx := o.ctx

	// Grab existing meta info.
	einfo, err := obs.GetInfo(meta.Name)
	if err != nil && err != ErrObjectNotFound {
		return nil, err
	}

	// Create a random subject prefixed with the object stream name.
	id := nuid.Next()
	chunkSubj := fmt.Sprintf(objChunksPreTmpl, obs.name, id)
	metaSubj := fmt.Sprintf(objMetaPreTmpl, obs.name, obj)

	// For async error handling
	var perr error
	var mu sync.Mutex
	setErr := func(err error) {
		mu.Lock()
		defer mu.Unlock()
		perr = err
	}
	getErr := func() error {
		mu.Lock()
		defer mu.Unlock()
		return perr
	}

	purgePartial := func() { obs.js.purgeStream(obs.stream, &streamPurgeRequest{Subject: chunkSubj}) }

	// Create our own JS context to handle errors etc.
	js, err := obs.js.nc.JetStream(PublishAsyncErrHandler(func(js JetStream, _ *Msg, err error) { setErr(err) }))
	if err != nil {
		return nil, err
	}

	chunkSize := objDefaultChunkSize
	if meta.Opts != nil && meta.Opts.ChunkSize > 0 {
		chunkSize = meta.Opts.ChunkSize
	}

	m, h := NewMsg(chunkSubj), sha256.New()
	chunk, sent, total := make([]byte, chunkSize), 0, uint64(0)
	info := &ObjectInfo{Bucket: obs.name, NUID: id, ObjectMeta: *meta}

	for r != nil {
		if ctx != nil {
			select {
			case <-ctx.Done():
				if ctx.Err() == context.Canceled {
					err = ctx.Err()
				} else {
					err = ErrTimeout
				}
			default:
			}
			if err != nil {
				purgePartial()
				return nil, err
			}
		}

		// Actual read.
		// TODO(dlc) - Deadline?
		n, err := r.Read(chunk)

		// EOF Processing.
		if err == io.EOF {
			// Finalize sha.
			sha := h.Sum(nil)
			// Place meta info.
			info.Size, info.Chunks = uint64(total), uint32(sent)
			info.Digest = fmt.Sprintf(objDigestTmpl, base64.URLEncoding.EncodeToString(sha[:]))
			break
		} else if err != nil {
			purgePartial()
			return nil, err
		}

		// Chunk processing.
		m.Data = chunk[:n]
		h.Write(m.Data)

		// Send msg itself.
		if _, err := js.PublishMsgAsync(m); err != nil {
			purgePartial()
			return nil, err
		}
		if err := getErr(); err != nil {
			purgePartial()
			return nil, err
		}
		// Update totals.
		sent++
		total += uint64(n)
	}

	// Publish the metadata.
	mm := NewMsg(metaSubj)
	mm.Header.Set(MsgRollup, MsgRollupSubject)
	mm.Data, err = json.Marshal(info)
	if err != nil {
		if r != nil {
			purgePartial()
		}
		return nil, err
	}
	// Send meta message.
	_, err = js.PublishMsgAsync(mm)
	if err != nil {
		if r != nil {
			purgePartial()
		}
		return nil, err
	}

	// Wait for all to be processed.
	select {
	case <-js.PublishAsyncComplete():
		if err := getErr(); err != nil {
			purgePartial()
			return nil, err
		}
	case <-time.After(obs.js.opts.wait):
		return nil, ErrTimeout
	}
	info.ModTime = time.Now().UTC()

	// Delete any original one.
	if einfo != nil && !einfo.Deleted {
		chunkSubj := fmt.Sprintf(objChunksPreTmpl, obs.name, einfo.NUID)
		obs.js.purgeStream(obs.stream, &streamPurgeRequest{Subject: chunkSubj})
	}

	return info, nil
}

// ObjectResult impl.
type objResult struct {
	sync.Mutex
	info *ObjectInfo
	r    io.ReadCloser
	err  error
	ctx  context.Context
}

func (info *ObjectInfo) isLink() bool {
	return info.ObjectMeta.Opts != nil && info.ObjectMeta.Opts.Link != nil
}

// Get will pull the object from the underlying stream.
func (obs *obs) Get(name string, opts ...ObjectOpt) (ObjectResult, error) {
	// Grab meta info.
	info, err := obs.GetInfo(name)
	if err != nil {
		return nil, err
	}
	if info.NUID == _EMPTY_ {
		return nil, ErrBadObjectMeta
	}

	// Check for object links.If single objects we do a pass through.
	if info.isLink() {
		if info.ObjectMeta.Opts.Link.Name == _EMPTY_ {
			return nil, errors.New("nats: link is a bucket")
		}
		lobs, err := obs.js.ObjectStore(info.ObjectMeta.Opts.Link.Bucket)
		if err != nil {
			return nil, err
		}
		return lobs.Get(info.ObjectMeta.Opts.Link.Name)
	}

	var o objOpts
	for _, opt := range opts {
		if opt != nil {
			if err := opt.configureObject(&o); err != nil {
				return nil, err
			}
		}
	}
	ctx := o.ctx

	result := &objResult{info: info, ctx: ctx}
	if info.Size == 0 {
		return result, nil
	}

	pr, pw := net.Pipe()
	result.r = pr

	gotErr := func(m *Msg, err error) {
		pw.Close()
		m.Sub.Unsubscribe()
		result.setErr(err)
	}

	// For calculating sum256
	h := sha256.New()

	processChunk := func(m *Msg) {
		if ctx != nil {
			select {
			case <-ctx.Done():
				if ctx.Err() == context.Canceled {
					err = ctx.Err()
				} else {
					err = ErrTimeout
				}
			default:
			}
			if err != nil {
				gotErr(m, err)
				return
			}
		}

		tokens, err := getMetadataFields(m.Reply)
		if err != nil {
			gotErr(m, err)
			return
		}

		// Write to our pipe.
		for b := m.Data; len(b) > 0; {
			n, err := pw.Write(b)
			if err != nil {
				gotErr(m, err)
				return
			}
			b = b[n:]
		}
		// Update sha256
		h.Write(m.Data)

		// Check if we are done.
		if tokens[ackNumPendingTokenPos] == objNoPending {
			pw.Close()
			m.Sub.Unsubscribe()

			// Make sure the digest matches.
			sha := h.Sum(nil)
			rsha, err := base64.URLEncoding.DecodeString(info.Digest)
			if err != nil {
				gotErr(m, err)
				return
			}
			if !bytes.Equal(sha[:], rsha) {
				gotErr(m, ErrDigestMismatch)
				return
			}
		}
	}

	chunkSubj := fmt.Sprintf(objChunksPreTmpl, obs.name, info.NUID)
	_, err = obs.js.Subscribe(chunkSubj, processChunk, OrderedConsumer())
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Delete will delete the object.
func (obs *obs) Delete(name string) error {
	// Grab meta info.
	info, err := obs.GetInfo(name)
	if err != nil {
		return err
	}
	if info.NUID == _EMPTY_ {
		return ErrBadObjectMeta
	}

	// Place a rollup delete marker.
	info.Deleted = true
	info.Size, info.Chunks, info.Digest = 0, 0, _EMPTY_

	metaSubj := fmt.Sprintf(objMetaPreTmpl, obs.name, sanitizeName(name))
	mm := NewMsg(metaSubj)
	mm.Data, err = json.Marshal(info)
	if err != nil {
		return err
	}
	mm.Header.Set(MsgRollup, MsgRollupSubject)
	_, err = obs.js.PublishMsg(mm)
	if err != nil {
		return err
	}

	// Purge chunks for the object.
	chunkSubj := fmt.Sprintf(objChunksPreTmpl, obs.name, info.NUID)
	return obs.js.purgeStream(obs.stream, &streamPurgeRequest{Subject: chunkSubj})
}

// AddLink will add a link to another object into this object store.
func (obs *obs) AddLink(name string, obj *ObjectInfo) (*ObjectInfo, error) {
	if obj == nil {
		return nil, errors.New("nats: object required")
	}
	if obj.Deleted {
		return nil, errors.New("nats: object is deleted")
	}
	name = sanitizeName(name)
	if !keyValid(name) {
		return nil, ErrInvalidObjectName
	}

	// Same object store.
	if obj.Bucket == obs.name {
		info := *obj
		info.Name = name
		if err := obs.UpdateMeta(obj.Name, &info.ObjectMeta); err != nil {
			return nil, err
		}
		return obs.GetInfo(name)
	}

	link := &ObjectLink{Bucket: obj.Bucket, Name: obj.Name}
	meta := &ObjectMeta{
		Name: name,
		Opts: &ObjectMetaOptions{Link: link},
	}
	return obs.Put(meta, nil)
}

// AddBucketLink will add a link to another object store.
func (ob *obs) AddBucketLink(name string, bucket ObjectStore) (*ObjectInfo, error) {
	if bucket == nil {
		return nil, errors.New("nats: bucket required")
	}
	name = sanitizeName(name)
	if !keyValid(name) {
		return nil, ErrInvalidObjectName
	}

	bos, ok := bucket.(*obs)
	if !ok {
		return nil, errors.New("nats: bucket malformed")
	}
	meta := &ObjectMeta{
		Name: name,
		Opts: &ObjectMetaOptions{Link: &ObjectLink{Bucket: bos.name}},
	}
	return ob.Put(meta, nil)
}

// PutBytes is convenience function to put a byte slice into this object store.
func (obs *obs) PutBytes(name string, data []byte, opts ...ObjectOpt) (*ObjectInfo, error) {
	return obs.Put(&ObjectMeta{Name: name}, bytes.NewReader(data), opts...)
}

// GetBytes is a convenience function to pull an object from this object store and return it as a byte slice.
func (obs *obs) GetBytes(name string, opts ...ObjectOpt) ([]byte, error) {
	result, err := obs.Get(name, opts...)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	var b bytes.Buffer
	if _, err := b.ReadFrom(result); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// PutBytes is convenience function to put a string into this object store.
func (obs *obs) PutString(name string, data string, opts ...ObjectOpt) (*ObjectInfo, error) {
	return obs.Put(&ObjectMeta{Name: name}, strings.NewReader(data), opts...)
}

// GetString is a convenience function to pull an object from this object store and return it as a string.
func (obs *obs) GetString(name string, opts ...ObjectOpt) (string, error) {
	result, err := obs.Get(name, opts...)
	if err != nil {
		return _EMPTY_, err
	}
	defer result.Close()

	var b bytes.Buffer
	if _, err := b.ReadFrom(result); err != nil {
		return _EMPTY_, err
	}
	return b.String(), nil
}

// PutFile is convenience function to put a file into an object store.
func (obs *obs) PutFile(file string, opts ...ObjectOpt) (*ObjectInfo, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return obs.Put(&ObjectMeta{Name: file}, f, opts...)
}

// GetFile is a convenience function to pull and object and place in a file.
func (obs *obs) GetFile(name, file string, opts ...ObjectOpt) error {
	// Expect file to be new.
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	result, err := obs.Get(name, opts...)
	if err != nil {
		os.Remove(f.Name())
		return err
	}
	defer result.Close()

	// Stream copy to the file.
	_, err = io.Copy(f, result)
	return err
}

// GetInfo will retrieve the current information for the object.
func (obs *obs) GetInfo(name string) (*ObjectInfo, error) {
	// Lookup the stream to get the bound subject.
	obj := sanitizeName(name)
	if !keyValid(obj) {
		return nil, ErrInvalidObjectName
	}

	// Grab last meta value we have.
	meta := fmt.Sprintf(objMetaPreTmpl, obs.name, obj)
	stream := fmt.Sprintf(objNameTmpl, obs.name)

	m, err := obs.js.GetLastMsg(stream, meta)
	if err != nil {
		if err == ErrMsgNotFound {
			err = ErrObjectNotFound
		}
		return nil, err
	}
	var info ObjectInfo
	if err := json.Unmarshal(m.Data, &info); err != nil {
		return nil, ErrBadObjectMeta
	}
	info.ModTime = m.Time
	return &info, nil
}

// UpdateMeta will update the meta data for the object.
func (obs *obs) UpdateMeta(name string, meta *ObjectMeta) error {
	if meta == nil {
		return ErrBadObjectMeta
	}
	// Grab meta info.
	info, err := obs.GetInfo(name)
	if err != nil {
		return err
	}
	// Copy new meta
	info.ObjectMeta = *meta
	mm := NewMsg(fmt.Sprintf(objMetaPreTmpl, obs.name, sanitizeName(meta.Name)))
	mm.Data, err = json.Marshal(info)
	if err != nil {
		return err
	}
	_, err = obs.js.PublishMsg(mm)
	return err
}

// Seal will seal the object store, no further modifications will be allowed.
func (obs *obs) Seal() error {
	stream := fmt.Sprintf(objNameTmpl, obs.name)
	si, err := obs.js.StreamInfo(stream)
	if err != nil {
		return err
	}
	// Seal the stream from being able to take on more messages.
	cfg := si.Config
	cfg.Sealed = true
	_, err = obs.js.UpdateStream(&cfg)
	return err
}

// Implementation for Watch
type objWatcher struct {
	updates chan *ObjectInfo
	sub     *Subscription
}

// Updates returns the interior channel.
func (w *objWatcher) Updates() <-chan *ObjectInfo {
	if w == nil {
		return nil
	}
	return w.updates
}

// Stop will unsubscribe from the watcher.
func (w *objWatcher) Stop() error {
	if w == nil {
		return nil
	}
	return w.sub.Unsubscribe()
}

// Watch for changes in the underlying store and receive meta information updates.
func (obs *obs) Watch(opts ...WatchOpt) (ObjectWatcher, error) {
	var o watchOpts
	for _, opt := range opts {
		if opt != nil {
			if err := opt.configureWatcher(&o); err != nil {
				return nil, err
			}
		}
	}

	var initDoneMarker bool

	w := &objWatcher{updates: make(chan *ObjectInfo, 32)}

	update := func(m *Msg) {
		var info ObjectInfo
		if err := json.Unmarshal(m.Data, &info); err != nil {
			return // TODO(dlc) - Communicate this upwards?
		}
		meta, err := m.Metadata()
		if err != nil {
			return
		}

		if !o.ignoreDeletes || !info.Deleted {
			info.ModTime = meta.Timestamp
			w.updates <- &info
		}

		if !initDoneMarker && meta.NumPending == 0 {
			initDoneMarker = true
			w.updates <- nil
		}
	}

	allMeta := fmt.Sprintf(objAllMetaPreTmpl, obs.name)
	_, err := obs.js.GetLastMsg(obs.stream, allMeta)
	if err == ErrMsgNotFound {
		initDoneMarker = true
		w.updates <- nil
	}

	// Used ordered consumer to deliver results.
	subOpts := []SubOpt{OrderedConsumer()}
	if !o.includeHistory {
		subOpts = append(subOpts, DeliverLastPerSubject())
	}
	sub, err := obs.js.Subscribe(allMeta, update, subOpts...)
	if err != nil {
		return nil, err
	}
	w.sub = sub
	return w, nil
}

// List will list all the objects in this store.
func (obs *obs) List(opts ...WatchOpt) ([]*ObjectInfo, error) {
	opts = append(opts, IgnoreDeletes())
	watcher, err := obs.Watch(opts...)
	if err != nil {
		return nil, err
	}
	defer watcher.Stop()

	var objs []*ObjectInfo
	for entry := range watcher.Updates() {
		if entry == nil {
			break
		}
		objs = append(objs, entry)
	}
	if len(objs) == 0 {
		return nil, ErrNoObjectsFound
	}
	return objs, nil
}

// ObjectBucketStatus  represents status of a Bucket, implements ObjectStoreStatus
type ObjectBucketStatus struct {
	nfo    *StreamInfo
	bucket string
}

// Bucket is the name of the bucket
func (s *ObjectBucketStatus) Bucket() string { return s.bucket }

// Description is the description supplied when creating the bucket
func (s *ObjectBucketStatus) Description() string { return s.nfo.Config.Description }

// TTL indicates how long objects are kept in the bucket
func (s *ObjectBucketStatus) TTL() time.Duration { return s.nfo.Config.MaxAge }

// Storage indicates the underlying JetStream storage technology used to store data
func (s *ObjectBucketStatus) Storage() StorageType { return s.nfo.Config.Storage }

// Replicas indicates how many storage replicas are kept for the data in the bucket
func (s *ObjectBucketStatus) Replicas() int { return s.nfo.Config.Replicas }

// Sealed indicates the stream is sealed and cannot be modified in any way
func (s *ObjectBucketStatus) Sealed() bool { return s.nfo.Config.Sealed }

// Size is the combined size of all data in the bucket including metadata, in bytes
func (s *ObjectBucketStatus) Size() uint64 { return s.nfo.State.Bytes }

// BackingStore indicates what technology is used for storage of the bucket
func (s *ObjectBucketStatus) BackingStore() string { return "JetStream" }

// StreamInfo is the stream info retrieved to create the status
func (s *ObjectBucketStatus) StreamInfo() *StreamInfo { return s.nfo }

// Status retrieves run-time status about a bucket
func (obs *obs) Status() (ObjectStoreStatus, error) {
	nfo, err := obs.js.StreamInfo(obs.stream)
	if err != nil {
		return nil, err
	}

	status := &ObjectBucketStatus{
		nfo:    nfo,
		bucket: obs.name,
	}

	return status, nil
}

// Read impl.
func (o *objResult) Read(p []byte) (n int, err error) {
	o.Lock()
	defer o.Unlock()
	if ctx := o.ctx; ctx != nil {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.Canceled {
				o.err = ctx.Err()
			} else {
				o.err = ErrTimeout
			}
		default:
		}
	}
	if o.err != nil {
		return 0, err
	}
	if o.r == nil {
		return 0, io.EOF
	}

	r := o.r.(net.Conn)
	r.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err = r.Read(p)
	if err, ok := err.(net.Error); ok && err.Timeout() {
		if ctx := o.ctx; ctx != nil {
			select {
			case <-ctx.Done():
				if ctx.Err() == context.Canceled {
					return 0, ctx.Err()
				} else {
					return 0, ErrTimeout
				}
			default:
				err = nil
			}
		}
	}
	return n, err
}

// Close impl.
func (o *objResult) Close() error {
	o.Lock()
	defer o.Unlock()
	if o.r == nil {
		return nil
	}
	return o.r.Close()
}

func (o *objResult) setErr(err error) {
	o.Lock()
	defer o.Unlock()
	o.err = err
}

func (o *objResult) Info() (*ObjectInfo, error) {
	o.Lock()
	defer o.Unlock()
	return o.info, o.err
}

func (o *objResult) Error() error {
	o.Lock()
	defer o.Unlock()
	return o.err
}
