# event-bench
1.) Clone repo locally

2.) `cd cmd`

3.) `go build`

4.) There are two ways to execute the program. To benchmark events generated without a binary value, simply call `./cmd`.
To generate a checksum for an event with a binary value, call `./cmd -b`

5.) Additional parameters:

param | type | description
------|------|---------------------
-i | int64 | Perform a number of iterations.
-b | bool | True = Generate binary events. False = Generate simple events.
-s | bool | True = When generating binary events, use 100kb image. False = 900kb image.
-p | bool | True = Print the encoded/decoded data.
-c | bool | True = Use contract client form of CBOR encoder/decoder. False = DSDK method/style.
