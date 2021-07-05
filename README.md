## Fast parameter discovery and attack tool.
![](pb.png?raw=true)


### Find 
Designed around [Arjun.py](https://github.com/s0md3v/Arjun), implemented in go with improved filtering capabilities.


#### Sequence:
- queries the page and searches for form values
- performs autocalibration (checks headers,status, response body) special filitering to filter reflected values (even if reflected multiple times)
- Splits the wordlist into chunks
- tries each chunk, if content changes it splits the chunk again, repeats until a final valid parameter is found.
  


### Attack

Not complete, starting this soon.