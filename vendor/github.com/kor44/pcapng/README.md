# PcapNG

This library provides packet reading capabilities of files in pcang format.


```go
// Create new reader:
f, _ := os.Open("file.pcapng")
defer f.Close()
r, err := NewReader(f)
data, ci, err := r.ReadPacketData()
```