Standard Package Status
-----------------------

Currently the standard packages that [pass their tests](https://github.com/tardisgo/tardisgo/blob/master/goroot/haxe/go1.4/src/haxetests.log) are shown below. If a standard package is not mentioned in the table below, please assume it does not work. 

The "testing" package is emulated in an ugly and part-working way. Packages "reflect", "os" & "syscall" are part-implemented, using an implementation of the nacl runtime.

Math-related packages may only fully work with cpp or js -D fullunsafe, partly due to modelling float32 as float64. 


| Name            | Passes Tests          | Comment                           |
| --------------- | --------------------- | --------------------------------- |
| archive         |                       |                                   |
| -- tar          |                       |                                   |
| -- zip          |                       |                                   |
| bufio           | yes                   |                                   |
| builtin         | no tests              | all built-in fns are implemented  |
| bytes           | yes                   |                                   |
| compress        |                       |                                   |
| -- bzip2        |                       |                                   |
| -- flate        |                       |                                   |
| -- gzip         |                       |                                   |
| -- lzw          |                       |                                   |
| -- zlib         |                       |                                   |
| container       |                       |                                   |
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
| database        |                       |                                   |
| -- sql          |                       |                                   |
| -- -- driver    | yes                   |                                   |
| debug           |                       |                                   |
| -- dwarf        |                       |                                   |
| -- elf          |                       |                                   |
| -- gosym        |                       |                                   |
| -- macho        |                       |                                   |
| -- pe           |                       |                                   |
| encoding        |                       |                                   |
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
| fmt             |                       | print side works, minor differences in type names |
| go              |                       |                                   |
| -- ast          |                       |                                   |
| -- build        |                       |                                   |
| -- doc          |                       |                                   |
| -- format       | yes                   |                                   |
| -- parser       |                       |                                   |
| -- printer      |                       |                                   |
| -- scanner      | yes                   |                                   |
| -- token        |                       |                                   |
| hash            |                       |                                   |
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
| index           |                       |                                   |
| -- suffixarray  | yes                   |                                   |
| io              |                       |                                   |
| -- ioutil       |                       |                                   |
| log             |                       |                                   |
| -- syslog       |                       |                                   |
| math            | yes                   |                                   |
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
| -- filepath     |                       |                                   |
| reflect         |                       | partially implemenented           |
| regexp          |                       |                                   |
| -- syntax       | yes                   |                                   |
| runtime         |                       | some general tests pass, NaN handled differently as a Map key |
| -- cgo          |                       | unsupported                       |
| -- debug        |                       | unsupported                       |
| -- pprof        |                       | unsupported                       |
| -- race         |                       | unsupported                       |
| sort            | yes                   |                                   |
| strconv         | yes                   |                                   |
| strings         | yes                   |                                   |
| sync            |                       |                                   |
| -- atomic       | yes                   |                                   |
| syscall         |                       | partial implementation via nacl   |
| testing         |                       | dummy at present                  |
| -- iotest       |                       |                                   |
| -- quick        |                       |                                   |
| text            |                       |                                   |
| -- scanner      | yes                   |                                   |
| -- tabwriter    | yes                   |                                   |
| -- template     |                       |                                   |
| -- -- parse     |                       |                                   |
| time            |                       |                                   |
| unicode         | yes                   |                                   |
| -- utf16        | yes                   |                                   |
| -- utf8         | yes                   |                                   |
| unsafe          | no tests              | pointer arithmetic unsupported    |

Score as at 6.3.15: 46/142 = 32%

(With thanks to [GopherJS](https://github.com/gopherjs/gopherjs/blob/master/doc/packages.md) for the layout above)
