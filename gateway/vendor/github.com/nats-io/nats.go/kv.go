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
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Notice: Experimental Preview
//
// This functionality is EXPERIMENTAL and may be changed in later releases.
type KeyValueManager interface {
	// KeyValue will lookup and bind to an existing KeyValue store.
	KeyValue(bucket string) (KeyValue, error)
	// CreateKeyValue will create a KeyValue store with the following configuration.
	CreateKeyValue(cfg *KeyValueConfig) (KeyValue, error)
	// DeleteKeyValue will delete this KeyValue store (JetStream stream).
	DeleteKeyValue(bucket string) error
}

// Notice: Experimental Preview
//
// This functionality is EXPERIMENTAL and may be changed in later releases.
type KeyValue interface {
	// Get returns the latest value for the key.
	Get(key string) (entry KeyValueEntry, err error)
	// GetRevision returns a specific revision value for the key.
	GetRevision(key string, revision uint64) (entry KeyValueEntry, err error)
	// Put will place the new value for the key into the store.
	Put(key string, value []byte) (revision uint64, err error)
	// PutString will place the string for the key into the store.
	PutString(key string, value string) (revision uint64, err error)
	// Create will add the key/value pair iff it does not exist.
	Create(key string, value []byte) (revision uint64, err error)
	// Update will update the value iff the latest revision matches.
	Update(key string, value []byte, last uint64) (revision uint64, err error)
	// Delete will place a delete marker and leave all revisions.
	Delete(key string, opts ...DeleteOpt) error
	// Purge will place a delete marker and remove all previous revisions.
	Purge(key string, opts ...DeleteOpt) error
	// Watch for any updates to keys that match the keys argument which could include wildcards.
	// Watch will send a nil entry when it has received all initial values.
	Watch(keys string, opts ...WatchOpt) (KeyWatcher, error)
	// WatchAll will invoke the callback for all updates.
	WatchAll(opts ...WatchOpt) (KeyWatcher, error)
	// Keys will return all keys.
	Keys(opts ...WatchOpt) ([]string, error)
	// History will return all historical values for the key.
	History(key string, opts ...WatchOpt) ([]KeyValueEntry, error)
	// Bucket returns the current bucket name.
	Bucket() string
	// PurgeDeletes will remove all current delete markers.
	PurgeDeletes(opts ...PurgeOpt) error
	// Status retrieves the status and configuration of a bucket
	Status() (KeyValueStatus, error)
}

// KeyValueStatus is run-time status about a Key-Value bucket
type KeyValueStatus interface {
	// Bucket the name of the bucket
	Bucket() string

	// Values is how many messages are in the bucket, including historical values
	Values() uint64

	// History returns the configured history kept per key
	History() int64

	// TTL is how long the bucket keeps values for
	TTL() time.Duration

	// BackingStore indicates what technology is used for storage of the bucket
	BackingStore() string
}

// KeyWatcher is what is returned when doing a watch.
type KeyWatcher interface {
	// Context returns watcher context optionally provided by nats.Context option.
	Context() context.Context
	// Updates returns a channel to read any updates to entries.
	Updates() <-chan KeyValueEntry
	// Stop will stop this watcher.
	Stop() error
}

type WatchOpt interface {
	configureWatcher(opts *watchOpts) error
}

// For nats.Context() support.
func (ctx ContextOpt) configureWatcher(opts *watchOpts) error {
	opts.ctx = ctx
	return nil
}

type watchOpts struct {
	ctx context.Context
	// Do not send delete markers to the update channel.
	ignoreDeletes bool
	// Include all history per subject, not just last one.
	includeHistory bool
	// retrieve only the meta data of the entry
	metaOnly bool
}

type watchOptFn func(opts *watchOpts) error

func (opt watchOptFn) configureWatcher(opts *watchOpts) error {
	return opt(opts)
}

// IncludeHistory instructs the key watcher to include historical values as well.
func IncludeHistory() WatchOpt {
	return watchOptFn(func(opts *watchOpts) error {
		opts.includeHistory = true
		return nil
	})
}

// IgnoreDeletes will have the key watcher not pass any deleted keys.
func IgnoreDeletes() WatchOpt {
	return watchOptFn(func(opts *watchOpts) error {
		opts.ignoreDeletes = true
		return nil
	})
}

// MetaOnly instructs the key watcher to retrieve only the entry meta data, not the entry value
func MetaOnly() WatchOpt {
	return watchOptFn(func(opts *watchOpts) error {
		opts.metaOnly = true
		return nil
	})
}

type PurgeOpt interface {
	configurePurge(opts *purgeOpts) error
}

type purgeOpts struct {
	dmthr time.Duration // Delete markers threshold
	ctx   context.Context
}

// DeleteMarkersOlderThan indicates that delete or purge markers older than that
// will be deleted as part of PurgeDeletes() operation, otherwise, only the data
// will be removed but markers that are recent will be kept.
// Note that if no option is specified, the default is 30 minutes. You can set
// this option to a negative value to instruct to always remove the markers,
// regardless of their age.
type DeleteMarkersOlderThan time.Duration

func (ttl DeleteMarkersOlderThan) configurePurge(opts *purgeOpts) error {
	opts.dmthr = time.Duration(ttl)
	return nil
}

// For nats.Context() support.
func (ctx ContextOpt) configurePurge(opts *purgeOpts) error {
	opts.ctx = ctx
	return nil
}

type DeleteOpt interface {
	configureDelete(opts *deleteOpts) error
}

type deleteOpts struct {
	// Remove all previous revisions.
	purge bool

	// Delete only if the latest revision matches.
	revision uint64
}

type deleteOptFn func(opts *deleteOpts) error

func (opt deleteOptFn) configureDelete(opts *deleteOpts) error {
	return opt(opts)
}

// LastRevision deletes if the latest revision matches.
func LastRevision(revision uint64) DeleteOpt {
	return deleteOptFn(func(opts *deleteOpts) error {
		opts.revision = revision
		return nil
	})
}

// purge removes all previous revisions.
func purge() DeleteOpt {
	return deleteOptFn(func(opts *deleteOpts) error {
		opts.purge = true
		return nil
	})
}

// KeyValueConfig is for configuring a KeyValue store.
type KeyValueConfig struct {
	Bucket       string
	Description  string
	MaxValueSize int32
	History      uint8
	TTL          time.Duration
	MaxBytes     int64
	Storage      StorageType
	Replicas     int
	Placement    *Placement
}

// Used to watch all keys.
const (
	KeyValueMaxHistory = 64
	AllKeys            = ">"
	kvLatestRevision   = 0
	kvop               = "KV-Operation"
	kvdel              = "DEL"
	kvpurge            = "PURGE"
)

type KeyValueOp uint8

const (
	KeyValuePut KeyValueOp = iota
	KeyValueDelete
	KeyValuePurge
)

func (op KeyValueOp) String() string {
	switch op {
	case KeyValuePut:
		return "KeyValuePutOp"
	case KeyValueDelete:
		return "KeyValueDeleteOp"
	case KeyValuePurge:
		return "KeyValuePurgeOp"
	default:
		return "Unknown Operation"
	}
}

// KeyValueEntry is a retrieved entry for Get or List or Watch.
type KeyValueEntry interface {
	// Bucket is the bucket the data was loaded from.
	Bucket() string
	// Key is the key that was retrieved.
	Key() string
	// Value is the retrieved value.
	Value() []byte
	// Revision is a unique sequence for this value.
	Revision() uint64
	// Created is the time the data was put in the bucket.
	Created() time.Time
	// Delta is distance from the latest value.
	Delta() uint64
	// Operation returns Put or Delete or Purge.
	Operation() KeyValueOp
}

// Errors
var (
	ErrKeyValueConfigRequired = errors.New("nats: config required")
	ErrInvalidBucketName      = errors.New("nats: invalid bucket name")
	ErrInvalidKey             = errors.New("nats: invalid key")
	ErrBucketNotFound         = errors.New("nats: bucket not found")
	ErrBadBucket              = errors.New("nats: bucket not valid key-value store")
	ErrKeyNotFound            = errors.New("nats: key not found")
	ErrKeyDeleted             = errors.New("nats: key was deleted")
	ErrHistoryToLarge         = errors.New("nats: history limited to a max of 64")
	ErrNoKeysFound            = errors.New("nats: no keys found")
)

const (
	kvBucketNameTmpl  = "KV_%s"
	kvSubjectsTmpl    = "$KV.%s.>"
	kvSubjectsPreTmpl = "$KV.%s."
	kvNoPending       = "0"
)

// Regex for valid keys and buckets.
var (
	validBucketRe = regexp.MustCompile(`\A[a-zA-Z0-9_-]+\z`)
	validKeyRe    = regexp.MustCompile(`\A[-/_=\.a-zA-Z0-9]+\z`)
)

// KeyValue will lookup and bind to an existing KeyValue store.
func (js *js) KeyValue(bucket string) (KeyValue, error) {
	if !js.nc.serverMinVersion(2, 6, 2) {
		return nil, errors.New("nats: key-value requires at least server version 2.6.2")
	}
	if !validBucketRe.MatchString(bucket) {
		return nil, ErrInvalidBucketName
	}
	stream := fmt.Sprintf(kvBucketNameTmpl, bucket)
	si, err := js.StreamInfo(stream)
	if err != nil {
		if err == ErrStreamNotFound {
			err = ErrBucketNotFound
		}
		return nil, err
	}
	// Do some quick sanity checks that this is a correctly formed stream for KV.
	// Max msgs per subject should be > 0.
	if si.Config.MaxMsgsPerSubject < 1 {
		return nil, ErrBadBucket
	}

	kv := &kvs{
		name:   bucket,
		stream: stream,
		pre:    fmt.Sprintf(kvSubjectsPreTmpl, bucket),
		js:     js,
		// Determine if we need to use the JS prefix in front of Put and Delete operations
		useJSPfx: js.opts.pre != defaultAPIPrefix,
	}
	return kv, nil
}

// CreateKeyValue will create a KeyValue store with the following configuration.
func (js *js) CreateKeyValue(cfg *KeyValueConfig) (KeyValue, error) {
	if !js.nc.serverMinVersion(2, 6, 2) {
		return nil, errors.New("nats: key-value requires at least server version 2.6.2")
	}
	if cfg == nil {
		return nil, ErrKeyValueConfigRequired
	}
	if !validBucketRe.MatchString(cfg.Bucket) {
		return nil, ErrInvalidBucketName
	}
	if _, err := js.AccountInfo(); err != nil {
		return nil, err
	}

	// Default to 1 for history. Max is 64 for now.
	history := int64(1)
	if cfg.History > 0 {
		if cfg.History > KeyValueMaxHistory {
			return nil, ErrHistoryToLarge
		}
		history = int64(cfg.History)
	}

	replicas := cfg.Replicas
	if replicas == 0 {
		replicas = 1
	}

	// We will set explicitly some values so that we can do comparison
	// if we get an "already in use" error and need to check if it is same.
	maxBytes := cfg.MaxBytes
	if maxBytes == 0 {
		maxBytes = -1
	}
	maxMsgSize := cfg.MaxValueSize
	if maxMsgSize == 0 {
		maxMsgSize = -1
	}
	// When stream's MaxAge is not set, server uses 2 minutes as the default
	// for the duplicate window. If MaxAge is set, and lower than 2 minutes,
	// then the duplicate window will be set to that. If MaxAge is greater,
	// we will cap the duplicate window to 2 minutes (to be consistent with
	// previous behavior).
	duplicateWindow := 2 * time.Minute
	if cfg.TTL > 0 && cfg.TTL < duplicateWindow {
		duplicateWindow = cfg.TTL
	}
	scfg := &StreamConfig{
		Name:              fmt.Sprintf(kvBucketNameTmpl, cfg.Bucket),
		Description:       cfg.Description,
		Subjects:          []string{fmt.Sprintf(kvSubjectsTmpl, cfg.Bucket)},
		MaxMsgsPerSubject: history,
		MaxBytes:          maxBytes,
		MaxAge:            cfg.TTL,
		MaxMsgSize:        maxMsgSize,
		Storage:           cfg.Storage,
		Replicas:          replicas,
		Placement:         cfg.Placement,
		AllowRollup:       true,
		DenyDelete:        true,
		Duplicates:        duplicateWindow,
		MaxMsgs:           -1,
		MaxConsumers:      -1,
	}

	// If we are at server version 2.7.2 or above use DiscardNew. We can not use DiscardNew for 2.7.1 or below.
	if js.nc.serverMinVersion(2, 7, 2) {
		scfg.Discard = DiscardNew
	}

	if _, err := js.AddStream(scfg); err != nil {
		// If we have a failure to add, it could be because we have
		// a config change if the KV was created against a pre 2.7.2
		// and we are now moving to a v2.7.2+. If that is the case
		// and the only difference is the discard policy, then update
		// the stream.
		if err == ErrStreamNameAlreadyInUse {
			if si, _ := js.StreamInfo(scfg.Name); si != nil {
				// To compare, make the server's stream info discard
				// policy same than ours.
				si.Config.Discard = scfg.Discard
				if reflect.DeepEqual(&si.Config, scfg) {
					_, err = js.UpdateStream(scfg)
				}
			}
		}
		if err != nil {
			return nil, err
		}
	}

	kv := &kvs{
		name:   cfg.Bucket,
		stream: scfg.Name,
		pre:    fmt.Sprintf(kvSubjectsPreTmpl, cfg.Bucket),
		js:     js,
		// Determine if we need to use the JS prefix in front of Put and Delete operations
		useJSPfx: js.opts.pre != defaultAPIPrefix,
	}
	return kv, nil
}

// DeleteKeyValue will delete this KeyValue store (JetStream stream).
func (js *js) DeleteKeyValue(bucket string) error {
	if !validBucketRe.MatchString(bucket) {
		return ErrInvalidBucketName
	}
	stream := fmt.Sprintf(kvBucketNameTmpl, bucket)
	return js.DeleteStream(stream)
}

type kvs struct {
	name   string
	stream string
	pre    string
	js     *js
	// If true, it means that APIPrefix/Domain was set in the context
	// and we need to add something to some of our high level protocols
	// (such as Put, etc..)
	useJSPfx bool
}

// Underlying entry.
type kve struct {
	bucket   string
	key      string
	value    []byte
	revision uint64
	delta    uint64
	created  time.Time
	op       KeyValueOp
}

func (e *kve) Bucket() string        { return e.bucket }
func (e *kve) Key() string           { return e.key }
func (e *kve) Value() []byte         { return e.value }
func (e *kve) Revision() uint64      { return e.revision }
func (e *kve) Created() time.Time    { return e.created }
func (e *kve) Delta() uint64         { return e.delta }
func (e *kve) Operation() KeyValueOp { return e.op }

func keyValid(key string) bool {
	if len(key) == 0 || key[0] == '.' || key[len(key)-1] == '.' {
		return false
	}
	return validKeyRe.MatchString(key)
}

// Get returns the latest value for the key.
func (kv *kvs) Get(key string) (KeyValueEntry, error) {
	e, err := kv.get(key, kvLatestRevision)
	if err != nil {
		if err == ErrKeyDeleted {
			return nil, ErrKeyNotFound
		}
		return nil, err
	}

	return e, nil
}

// GetRevision returns a specific revision value for the key.
func (kv *kvs) GetRevision(key string, revision uint64) (KeyValueEntry, error) {
	e, err := kv.get(key, revision)
	if err != nil {
		if err == ErrKeyDeleted {
			return nil, ErrKeyNotFound
		}
		return nil, err
	}

	return e, nil
}

func (kv *kvs) get(key string, revision uint64) (KeyValueEntry, error) {
	if !keyValid(key) {
		return nil, ErrInvalidKey
	}

	var b strings.Builder
	b.WriteString(kv.pre)
	b.WriteString(key)

	var m *RawStreamMsg
	var err error
	if revision == kvLatestRevision {
		m, err = kv.js.GetLastMsg(kv.stream, b.String())
	} else {
		m, err = kv.js.GetMsg(kv.stream, revision)
		if err == nil && m.Subject != b.String() {
			return nil, ErrKeyNotFound
		}
	}

	if err != nil {
		if err == ErrMsgNotFound {
			err = ErrKeyNotFound
		}
		return nil, err
	}

	entry := &kve{
		bucket:   kv.name,
		key:      key,
		value:    m.Data,
		revision: m.Sequence,
		created:  m.Time,
	}

	// Double check here that this is not a DEL Operation marker.
	if len(m.Header) > 0 {
		switch m.Header.Get(kvop) {
		case kvdel:
			entry.op = KeyValueDelete
			return entry, ErrKeyDeleted
		case kvpurge:
			entry.op = KeyValuePurge
			return entry, ErrKeyDeleted
		}
	}

	return entry, nil
}

// Put will place the new value for the key into the store.
func (kv *kvs) Put(key string, value []byte) (revision uint64, err error) {
	if !keyValid(key) {
		return 0, ErrInvalidKey
	}

	var b strings.Builder
	if kv.useJSPfx {
		b.WriteString(kv.js.opts.pre)
	}
	b.WriteString(kv.pre)
	b.WriteString(key)

	pa, err := kv.js.Publish(b.String(), value)
	if err != nil {
		return 0, err
	}
	return pa.Sequence, err
}

// PutString will place the string for the key into the store.
func (kv *kvs) PutString(key string, value string) (revision uint64, err error) {
	return kv.Put(key, []byte(value))
}

// Create will add the key/value pair iff it does not exist.
func (kv *kvs) Create(key string, value []byte) (revision uint64, err error) {
	v, err := kv.Update(key, value, 0)
	if err == nil {
		return v, nil
	}

	// TODO(dlc) - Since we have tombstones for DEL ops for watchers, this could be from that
	// so we need to double check.
	if e, err := kv.get(key, kvLatestRevision); err == ErrKeyDeleted {
		return kv.Update(key, value, e.Revision())
	}

	return 0, err
}

// Update will update the value iff the latest revision matches.
func (kv *kvs) Update(key string, value []byte, revision uint64) (uint64, error) {
	if !keyValid(key) {
		return 0, ErrInvalidKey
	}

	var b strings.Builder
	if kv.useJSPfx {
		b.WriteString(kv.js.opts.pre)
	}
	b.WriteString(kv.pre)
	b.WriteString(key)

	m := Msg{Subject: b.String(), Header: Header{}, Data: value}
	m.Header.Set(ExpectedLastSubjSeqHdr, strconv.FormatUint(revision, 10))

	pa, err := kv.js.PublishMsg(&m)
	if err != nil {
		return 0, err
	}
	return pa.Sequence, err
}

// Delete will place a delete marker and leave all revisions.
func (kv *kvs) Delete(key string, opts ...DeleteOpt) error {
	if !keyValid(key) {
		return ErrInvalidKey
	}

	var b strings.Builder
	if kv.useJSPfx {
		b.WriteString(kv.js.opts.pre)
	}
	b.WriteString(kv.pre)
	b.WriteString(key)

	// DEL op marker. For watch functionality.
	m := NewMsg(b.String())

	var o deleteOpts
	for _, opt := range opts {
		if opt != nil {
			if err := opt.configureDelete(&o); err != nil {
				return err
			}
		}
	}

	if o.purge {
		m.Header.Set(kvop, kvpurge)
		m.Header.Set(MsgRollup, MsgRollupSubject)
	} else {
		m.Header.Set(kvop, kvdel)
	}

	if o.revision != 0 {
		m.Header.Set(ExpectedLastSubjSeqHdr, strconv.FormatUint(o.revision, 10))
	}

	_, err := kv.js.PublishMsg(m)
	return err
}

// Purge will remove the key and all revisions.
func (kv *kvs) Purge(key string, opts ...DeleteOpt) error {
	return kv.Delete(key, append(opts, purge())...)
}

const kvDefaultPurgeDeletesMarkerThreshold = 30 * time.Minute

// PurgeDeletes will remove all current delete markers.
// This is a maintenance option if there is a larger buildup of delete markers.
// See DeleteMarkersOlderThan() option for more information.
func (kv *kvs) PurgeDeletes(opts ...PurgeOpt) error {
	var o purgeOpts
	for _, opt := range opts {
		if opt != nil {
			if err := opt.configurePurge(&o); err != nil {
				return err
			}
		}
	}
	// Transfer possible context purge option to the watcher. This is the
	// only option that matters for the PurgeDeletes() feature.
	var wopts []WatchOpt
	if o.ctx != nil {
		wopts = append(wopts, Context(o.ctx))
	}
	watcher, err := kv.WatchAll(wopts...)
	if err != nil {
		return err
	}
	defer watcher.Stop()

	var limit time.Time
	olderThan := o.dmthr
	// Negative value is used to instruct to always remove markers, regardless
	// of age. If set to 0 (or not set), use our default value.
	if olderThan == 0 {
		olderThan = kvDefaultPurgeDeletesMarkerThreshold
	}
	if olderThan > 0 {
		limit = time.Now().Add(-olderThan)
	}

	var deleteMarkers []KeyValueEntry
	for entry := range watcher.Updates() {
		if entry == nil {
			break
		}
		if op := entry.Operation(); op == KeyValueDelete || op == KeyValuePurge {
			deleteMarkers = append(deleteMarkers, entry)
		}
	}

	var (
		pr streamPurgeRequest
		b  strings.Builder
	)
	// Do actual purges here.
	for _, entry := range deleteMarkers {
		b.WriteString(kv.pre)
		b.WriteString(entry.Key())
		pr.Subject = b.String()
		pr.Keep = 0
		if olderThan > 0 && entry.Created().After(limit) {
			pr.Keep = 1
		}
		if err := kv.js.purgeStream(kv.stream, &pr); err != nil {
			return err
		}
		b.Reset()
	}
	return nil
}

// Keys() will return all keys.
func (kv *kvs) Keys(opts ...WatchOpt) ([]string, error) {
	opts = append(opts, IgnoreDeletes(), MetaOnly())
	watcher, err := kv.WatchAll(opts...)
	if err != nil {
		return nil, err
	}
	defer watcher.Stop()

	var keys []string
	for entry := range watcher.Updates() {
		if entry == nil {
			break
		}
		keys = append(keys, entry.Key())
	}
	if len(keys) == 0 {
		return nil, ErrNoKeysFound
	}
	return keys, nil
}

// History will return all values for the key.
func (kv *kvs) History(key string, opts ...WatchOpt) ([]KeyValueEntry, error) {
	opts = append(opts, IncludeHistory())
	watcher, err := kv.Watch(key, opts...)
	if err != nil {
		return nil, err
	}
	defer watcher.Stop()

	var entries []KeyValueEntry
	for entry := range watcher.Updates() {
		if entry == nil {
			break
		}
		entries = append(entries, entry)
	}
	if len(entries) == 0 {
		return nil, ErrKeyNotFound
	}
	return entries, nil
}

// Implementation for Watch
type watcher struct {
	mu          sync.Mutex
	updates     chan KeyValueEntry
	sub         *Subscription
	initDone    bool
	initPending uint64
	received    uint64
	ctx         context.Context
}

// Context returns the context for the watcher if set.
func (w *watcher) Context() context.Context {
	if w == nil {
		return nil
	}
	return w.ctx
}

// Updates returns the interior channel.
func (w *watcher) Updates() <-chan KeyValueEntry {
	if w == nil {
		return nil
	}
	return w.updates
}

// Stop will unsubscribe from the watcher.
func (w *watcher) Stop() error {
	if w == nil {
		return nil
	}
	return w.sub.Unsubscribe()
}

// WatchAll watches all keys.
func (kv *kvs) WatchAll(opts ...WatchOpt) (KeyWatcher, error) {
	return kv.Watch(AllKeys, opts...)
}

// Watch will fire the callback when a key that matches the keys pattern is updated.
// keys needs to be a valid NATS subject.
func (kv *kvs) Watch(keys string, opts ...WatchOpt) (KeyWatcher, error) {
	var o watchOpts
	for _, opt := range opts {
		if opt != nil {
			if err := opt.configureWatcher(&o); err != nil {
				return nil, err
			}
		}
	}

	// Could be a pattern so don't check for validity as we normally do.
	var b strings.Builder
	b.WriteString(kv.pre)
	b.WriteString(keys)
	keys = b.String()

	// We will block below on placing items on the chan. That is by design.
	w := &watcher{updates: make(chan KeyValueEntry, 256), ctx: o.ctx}

	update := func(m *Msg) {
		tokens, err := getMetadataFields(m.Reply)
		if err != nil {
			return
		}
		if len(m.Subject) <= len(kv.pre) {
			return
		}
		subj := m.Subject[len(kv.pre):]

		var op KeyValueOp
		if len(m.Header) > 0 {
			switch m.Header.Get(kvop) {
			case kvdel:
				op = KeyValueDelete
			case kvpurge:
				op = KeyValuePurge
			}
		}
		delta := uint64(parseNum(tokens[ackNumPendingTokenPos]))
		w.mu.Lock()
		defer w.mu.Unlock()
		if !o.ignoreDeletes || (op != KeyValueDelete && op != KeyValuePurge) {
			entry := &kve{
				bucket:   kv.name,
				key:      subj,
				value:    m.Data,
				revision: uint64(parseNum(tokens[ackStreamSeqTokenPos])),
				created:  time.Unix(0, parseNum(tokens[ackTimestampSeqTokenPos])),
				delta:    delta,
				op:       op,
			}
			w.updates <- entry
		}
		// Check if done and initial values.
		if !w.initDone {
			w.received++
			// We set this on the first trip through..
			if w.initPending == 0 {
				w.initPending = delta
			}
			if w.received > w.initPending || delta == 0 {
				w.initDone = true
				w.updates <- nil
			}
		}
	}

	// Used ordered consumer to deliver results.
	subOpts := []SubOpt{OrderedConsumer()}
	if !o.includeHistory {
		subOpts = append(subOpts, DeliverLastPerSubject())
	}
	if o.metaOnly {
		subOpts = append(subOpts, HeadersOnly())
	}
	if o.ctx != nil {
		subOpts = append(subOpts, Context(o.ctx))
	}
	// Create the sub and rest of initialization under the lock.
	// We want to prevent the race between this code and the
	// update() callback.
	w.mu.Lock()
	defer w.mu.Unlock()
	sub, err := kv.js.Subscribe(keys, update, subOpts...)
	if err != nil {
		return nil, err
	}
	sub.mu.Lock()
	// If there were no pending messages at the time of the creation
	// of the consumer, send the marker.
	if sub.jsi != nil && sub.jsi.pending == 0 {
		w.initDone = true
		w.updates <- nil
	}
	// Set us up to close when the waitForMessages func returns.
	sub.pDone = func() {
		close(w.updates)
	}
	sub.mu.Unlock()

	w.sub = sub
	return w, nil
}

// Bucket returns the current bucket name (JetStream stream).
func (kv *kvs) Bucket() string {
	return kv.name
}

// KeyValueBucketStatus represents status of a Bucket, implements KeyValueStatus
type KeyValueBucketStatus struct {
	nfo    *StreamInfo
	bucket string
}

// Bucket the name of the bucket
func (s *KeyValueBucketStatus) Bucket() string { return s.bucket }

// Values is how many messages are in the bucket, including historical values
func (s *KeyValueBucketStatus) Values() uint64 { return s.nfo.State.Msgs }

// History returns the configured history kept per key
func (s *KeyValueBucketStatus) History() int64 { return s.nfo.Config.MaxMsgsPerSubject }

// TTL is how long the bucket keeps values for
func (s *KeyValueBucketStatus) TTL() time.Duration { return s.nfo.Config.MaxAge }

// BackingStore indicates what technology is used for storage of the bucket
func (s *KeyValueBucketStatus) BackingStore() string { return "JetStream" }

// StreamInfo is the stream info retrieved to create the status
func (s *KeyValueBucketStatus) StreamInfo() *StreamInfo { return s.nfo }

// Status retrieves the status and configuration of a bucket
func (kv *kvs) Status() (KeyValueStatus, error) {
	nfo, err := kv.js.StreamInfo(kv.stream)
	if err != nil {
		return nil, err
	}

	return &KeyValueBucketStatus{nfo: nfo, bucket: kv.name}, nil
}
