## Fast parameter discovery and attack tool.
![](pb.png?raw=true)


```
Usage: parabuster [mode] [options]

Modes:
     find                 Discovers paramaters for an URL.
     attack               Fuzzes known parameters for issues.

Usage of find:
  -chunk int
        Chunk Size (default 50)
  -method string
        Method [get,post,all] (default "all")
  -threads int
        Concurent threads (default 10)
  -url string
        Target URL to test
  -wordlist string
        Parameter wordlist

Usage of attack:
  -payloads string
        File containing payloads
  -url string
        Target URL to test
```

### Find 
Designed around [Arjun.py](https://github.com/s0md3v/Arjun), implemented in go with improved filtering capabilities. 


#### Sequence:
- queries the page and searches for form values, if found they are added to the wordlist queue.
- performs autocalibration, checks headers, status, response body and applies special filitering to filter reflected values (even if reflected multiple times)
- Splits the wordlist into chunks
- tries each chunk, if content changes it splits the chunk again, repeats until a final valid parameter is found.
  


### Attack

Not complete, starting this soon.