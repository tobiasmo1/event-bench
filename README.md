# event-bench
1.) Clone repo locally

2.) `cd cmd`

3.) `go build`

4.) Execute the program, for example
`./cmd -i=1000 -b=true -p=false -c=true -s=false`

5.) Additional parameters:

param | type | description
------|------|---------------------
-i | int64 | Perform a number of iterations.
-b | bool | Generate binary events. True=binary, False=simple.
-p | bool | Print encoded/decoded data and generated checksum (incurs I/O overhead). True=print, False=silent.
-c | bool | Use CBOR encoder/decoder in form provided in core-contract. True=contract's style, False=DSDK method/style.
-s | bool | When generating binary events, use Small image. True=100kb image, False=900kb image.
