Standard Package Status
-----------------------

The standard packages that [pass their tests](https://github.com/tardisgo/tardisgo/blob/master/goroot/haxe/go1.4/src/tgotests.log) in js, c#, c++, or java are shown below. Only the js target has been systematically tested for all packages. 

The "testing" package is emulated in an ugly and part-working way, all tests are run in Short mode. Packages "reflect", "os" & "syscall" are part-implemented, using an implementation of the nacl runtime.

| Name            | Passes tests in?      | Comment                           |
| --------------- | --------------------- | --------------------------------- |
| archive         | no code               |                                   |
| -- tar          | c++, js               | c#/java: issue relating to incorrect map handling in reflect |
| -- zip          | c#, java, js          | c++: hangs (or takes too long)    |
| bufio           | c++, js               | c#/java: endless loop in ReadByte?     |
| builtin         | no tests              | all built-in functions are implemented |
| bytes           | c++, c#, java, js     |                                   |
| compress        | no code               |                                   |
| -- bzip2        | c++, c#, java, js     |                                   |
| -- flate        | c++, c#, java, js     |                                   |
| -- gzip         | c++, java, js         | c#: Copy hung                     |
| -- lzw          | c++, c#, java, js     |                                   |
| -- zlib         | c++, c#, java, js     |                                   |
| container       | no code               |                                   |
| -- heap         | c++, c#, java, js     |                                   |
| -- list         | c++, c#, java, js     |                                   |
| -- ring         | c++, c#, java, js     |                                   |
| crypto          | no tests              |                                   |
| -- aes          | js                    |                                   |
| -- cipher       | js                    |                                   |
| -- des          | js                    |                                   |
| -- dsa          | js                    |                                   |
| -- ecdsa        | js                    |                                   | 
| -- elliptic     | js                    |                                   |
| -- hmac         | js                    |                                   |
| -- md5          | js                    |                                   |
| -- rand         | js                    |                                   |
| -- rc4          | js                    |                                   |
| -- rsa          |                       | waiting for reflect.Call          |
| -- sha1         | js                    |                                   |
| -- sha256       | js                    |                                   |
| -- sha512       | js                    |                                   |
| -- subtle       |                       | waiting for reflect.Call          |
| -- tls          |                       | panic: duplicate function name: crypto/tls.run$1 |
| -- x509         |                       | modified tests (as for Windows) pass on js, but tests take >30 mins to run on nodejs, so not marked as working |
| -- -- pkix      | no tests              |                                   |
| database        | no code               |                                   |
| -- sql          |                       | panic: duplicate function name: database/sql.Query$1 |
| -- -- driver    | js                    |                                   |
| debug           | no code               |                                   |
| -- dwarf        |                       | interface type assert failed      |
| -- elf          | js                    |                                   |
| -- gosym        | js                    |                                   |
| -- macho        |                       | file_test.go 169: duplicate architecture |
| -- pe           | js                    |                                   |
| -- plan9obj     | js                    |                                   |
| encoding        | no tests              |                                   |
| -- ascii85      | js                    |                                   |
| -- asn1         |                       | 2 errors, probably both UTF-8 encoding related |
| -- base32       | js                    |                                   |
| -- base64       | js                    |                                   |
| -- binary       |                       | reflect: unknown method using value obtained using unexported field |
| -- csv          | js                    |                                   |
| -- gob          |                       | fatal error: stack overflow       |
| -- hex          | js                    |                                   |
| -- json         |                       | 2 errors related to seeing fields |
| -- pem          | js                    |                                   |
| -- xml          |                       | multiple errors, then crashes     |
| errors          | c++, c#, java, js     |                                   |
| expvar          |                       | Haxe try-catch exception after JSON unmarshall |
| flag            | js                    | but no way to pass flags in yet   |
| fmt             | c++, js               | minor differences in type names, c#/java: type conversion error in reflect package |
| go              | no code               |                                   |
| -- ast          |                       | multiple errors                   |
| -- build        |                       | $GOROOT/$GOPATH not set           |
| -- doc          |                       | os.PathError, probably test-data set-up issue |
| -- format       | js                    |                                   |
| -- parser       |                       | os.PathError, probably test-data set-up issue |
| -- printer      |                       | os.PathError, probably test-data set-up issue |
| -- scanner      | js                    |                                   |
| -- token        | js                    |                                   |
| hash            | no tests              |                                   |
| -- adler32      | js                    |                                   |
| -- crc32        | js                    |                                   |
| -- crc64        | js                    |                                   |
| -- fnv          | js                    |                                   |
| html            | js                    |                                   |
| -- template     |                       | waiting for reflect.Call          |
| image           | js                    |                                   |
| -- color        | js                    |                                   |
| -- -- palette   | no tests              |                                   |
| -- draw         | js                    |                                   |
| -- gif          | js                    |                                   |
| -- jpeg         | js                    |                                   |
| -- png          |                       | concrete type assert failed       |
| index           | no code               |                                   |
| -- suffixarray  | js                    |                                   |
| io              | c++, c#, java, js     |                                   |
| -- ioutil       | c++, c#, java, js     |                                   |
| log             |                       | multiple matching errors          |
| -- syslog       | no tests              |                                   |
| math            | c++, js               | c#/java: float32/int overflow issues |
| -- big          |                       | waiting for reflect.Call          |
| -- cmplx        | c++, c#, java, js     |                                   |
| -- rand         |                       | waiting for reflect.Method        |
| mime            | js                    |                                   |
| -- multipart    |                       | hangs                             |
| net             |                       | hangs                             |
| -- http         |                       | M not declared by dummy package testing |
| -- -- cgi       | no                    | fork/exec not implemented         |
| -- -- cookiejar |                       | errors in TestUpdateAndDelete     |
| -- -- fcgi      | js                    |                                   |
| -- -- httptest  |                       | hangs                             |
| -- -- httputil  |                       | hangs                             |
| -- -- internal  | js                    |                                   |
| -- -- pprof     | no tests              |                                   |
| -- mail         | js                    |                                   |
| -- rpc          |                       | hangs                             |
| -- -- jsonrpc   |                       | hangs                             |
| -- smtp         |                       | panic: syscall.stopTimer()        |
| -- textproto    | js                    |                                   |
| -- url          | js                    |                                   |
| os              | js                    | passes modified tests (no system files to read) |
| -- exec         | -                     | tests fail, dummy testing T.Skip() not properly implemented |
| -- signal       | -                     | no tests (for nacl)               |
| -- user         | -                     | tests run with (correct) errors   |
| path            | js                    |                                   |
| -- filepath     | js                    |                                   |
| reflect         |                       | partially implemenented - 1st error: invalid function reference |
| regexp          | js                    |                                   |
| -- syntax       | js                    |                                   |
| runtime         | some                  | some general tests pass, NaN handled differently as a Map key |
| -- cgo          | -                     | unsupported                       |
| -- debug        | -                     | unsupported                       |
| -- pprof        | -                     | unsupported                       |
| -- race         | -                     | unsupported                       |
| sort            | c++, c#, java, js     |                                   |
| strconv         | c++, js               | c#/java: float32 issues           |
| strings         | c++, c#, java, js     |                                   |
| sync            |                       | hangs                             |
| -- atomic       | js                    |                                   |
| syscall         |                       | partial implementation via nacl   |
| testing         |                       | dummy at present                  |
| -- iotest       |                       | dummy                             |
| -- quick        |                       | dummy                             |
| text            | no code               |                                   |
| -- scanner      | js                    |                                   |
| -- tabwriter    | js                    |                                   |
| -- template     |                       | hangs                             |
| -- -- parse     |                       | 2 errors related to integer 1e19  |
| time            |                       | duration error and waiting for reflect.Call |
| unicode         | c++, c#, java, js     |                                   |
| -- utf16        | c++, c#, java, js     |                                   |
| -- utf8         | c++, c#, java, js     |                                   |
| unsafe          | no tests              | pointer arithmetic unsupported, but other functionalty should work |

(With thanks to [GopherJS](https://github.com/gopherjs/gopherjs/blob/master/doc/packages.md) for the layout above)
