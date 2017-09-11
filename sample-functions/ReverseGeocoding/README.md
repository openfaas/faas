### ReverseGeocoding

Get address from GPS coordinates

### Usage

Once the function is created, and the service name ReverseGeocoding, it can be called like the following:

```
curl -d 'LONGITUDE LATITUDE' http://localhost:8080/function/ReverseGeocoding
```

### Example

```
$ curl -d '7.116703 43.581915' http://localhost:8080/function/ReverseGeocoding
20 Avenue Reibaud, 06600 Antibes, France
```
