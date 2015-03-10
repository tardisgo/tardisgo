Standard Package Status
-----------------------

The standard packages that [pass their tests](https://github.com/tardisgo/tardisgo/blob/master/goroot/haxe/go1.4/src/haxetests.log) are shown below, those without any comment have not yet passed their tests. 

The "testing" package is emulated in an ugly and part-working way. Packages "reflect", "os" & "syscall" are part-implemented, using an implementation of the nacl runtime.

Math-related packages may only fully work with cpp or js -D fullunsafe, partly due to modelling float32 as float64. 


| Name            | Passes Tests          | Comment                           |
| --------------- | --------------------- | --------------------------------- |
| archive         | no code               |                                   |
| -- tar          | yes                   |                                   |
| -- zip          | yes                   |                                   |
| bufio           | yes                   |                                   |
| builtin         | no tests              | all built-in functions are implemented |
| bytes           | yes                   |                                   |
| compress        | no code               |                                   |
| -- bzip2        | yes                   |                                   |
| -- flate        | yes                   |                                   |
| -- gzip         | yes                   |                                   |
| -- lzw          | yes                   |                                   |
| -- zlib         | yes                   |                                   |
| container       | no code               |                                   |
| -- heap         | yes                   |                                   |
| -- list         | yes                   |                                   |
| -- ring         | yes                   |                                   |
| crypto          |                       |                                   |
| -- aes          | yes                   |                                   |
| -- cipher       | yes                   |                                   |
| -- des          | yes                   |                                   |
| -- dsa          |                       |                                   |
| -- ecdsa        |                       |                                   |
| -- elliptic     |                       |                                   |
| -- hmac         |                       |                                   |
| -- md5          | yes                   |                                   |
| -- rand         |                       |                                   |
| -- rc4          | yes                   |                                   |
| -- rsa          |                       |                                   |
| -- sha1         | yes                   |                                   |
| -- sha256       | yes                   |                                   |
| -- sha512       | yes                   |                                   |
| -- subtle       |                       |                                   |
| -- tls          |                       |                                   |
| -- x509         |                       |                                   |
| -- -- pkix      |                       |                                   |
| database        | no code               |                                   |
| -- sql          |                       |                                   |
| -- -- driver    | yes                   |                                   |
| debug           | no code               |                                   |
| -- dwarf        |                       |                                   |
| -- elf          |                       |                                   |
| -- gosym        |                       |                                   |
| -- macho        |                       |                                   |
| -- pe           |                       |                                   |
| encoding        | no tests              |                                   |
| -- ascii85      | yes                   |                                   |
| -- asn1         |                       |                                   |
| -- base32       | yes                   |                                   |
| -- base64       | yes                   |                                   |
| -- binary       |                       |                                   |
| -- csv          | yes                   |                                   |
| -- gob          |                       |                                   |
| -- hex          | yes                   |                                   |
| -- json         |                       |                                   |
| -- pem          |                       |                                   |
| -- xml          |                       |                                   |
| errors          | yes                   |                                   |
| expvar          |                       |                                   |
| flag            | yes                   | but no way to pass flags in yet   |
| fmt             | some                  | print side works, minor differences in type names |
| go              | no code               |                                   |
| -- ast          |                       |                                   |
| -- build        |                       |                                   |
| -- doc          |                       |                                   |
| -- format       | yes                   |                                   |
| -- parser       |                       |                                   |
| -- printer      |                       |                                   |
| -- scanner      | yes                   |                                   |
| -- token        |                       |                                   |
| hash            | no tests              |                                   |
| -- adler32      | yes                   |                                   |
| -- crc32        | yes                   |                                   |
| -- crc64        | yes                   |                                   |
| -- fnv          | yes                   |                                   |
| html            | yes                   |                                   |
| -- template     |                       |                                   |
| image           |                       |                                   |
| -- color        | yes                   |                                   |
| -- -- palette   |                       |                                   |
| -- draw         | yes                   |                                   |
| -- gif          |                       |                                   |
| -- jpeg         |                       |                                   |
| -- png          |                       |                                   |
| index           | no code               |                                   |
| -- suffixarray  | yes                   |                                   |
| io              | yes                   |                                   |
| -- ioutil       | yes                   |                                   |
| log             |                       |                                   |
| -- syslog       |                       |                                   |
| math            | yes                   | requires fullunsafe mode in js to pass all tests |
| -- big          |                       |                                   |
| -- cmplx        | yes                   |                                   |
| -- rand         |                       |                                   |
| mime            |                       |                                   |
| -- multipart    |                       |                                   |
| net             |                       |                                   |
| -- http         |                       |                                   |
| -- -- cgi       |                       |                                   |
| -- -- cookiejar |                       |                                   |
| -- -- fcgi      |                       |                                   |
| -- -- httptest  |                       |                                   |
| -- -- httputil  |                       |                                   |
| -- -- pprof     |                       |                                   |
| -- mail         |                       |                                   |
| -- rpc          |                       |                                   |
| -- -- jsonrpc   |                       |                                   |
| -- smtp         |                       |                                   |
| -- textproto    |                       |                                   |
| -- url          |                       |                                   |
| os              |                       |                                   |
| -- exec         |                       |                                   |
| -- signal       |                       |                                   |
| -- user         |                       | will return an error if called    |
| path            | yes                   |                                   |
| -- filepath     | yes                   |                                   |
| reflect         |                       | partially implemenented           |
| regexp          | yes                   | requires fullunsafe mode in js to pass all tests |
| -- syntax       | yes                   |                                   |
| runtime         | some                  | some general tests pass, NaN handled differently as a Map key |
| -- cgo          | -                     | unsupported                       |
| -- debug        | -                     | unsupported                       |
| -- pprof        | -                     | unsupported                       |
| -- race         | -                     | unsupported                       |
| sort            | yes                   |                                   |
| strconv         | yes                   | only fully passes in C++ due to float -0 issues |
| strings         | yes                   |                                   |
| sync            |                       |                                   |
| -- atomic       | yes                   |                                   |
| syscall         |                       | partial implementation via nacl   |
| testing         |                       | dummy at present                  |
| -- iotest       |                       |                                   |
| -- quick        |                       |                                   |
| text            | no code               |                                   |
| -- scanner      | yes                   |                                   |
| -- tabwriter    | yes                   |                                   |
| -- template     |                       |                                   |
| -- -- parse     |                       |                                   |
| time            |                       |                                   |
| unicode         | yes                   |                                   |
| -- utf16        | yes                   |                                   |
| -- utf8         | yes                   |                                   |
| unsafe          | no tests              | pointer arithmetic unsupported, but other functionalty should work |


TODO (requires serious auomation): add library status for each target

(With thanks to [GopherJS](https://github.com/gopherjs/gopherjs/blob/master/doc/packages.md) for the layout above)
