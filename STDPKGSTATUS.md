Standard Package Status
-----------------------

The standard packages that [pass their tests](https://github.com/tardisgo/tardisgo/blob/master/goroot/haxe/go1.4/src/tgotests.log) in js, c#, c++, or java are shown below. (Entries in brackets mean that the tests pass, but take too long to be included in the automated regression testing.)

The "testing" package is emulated in an ugly and part-working way, all tests are run in Short mode. Packages "reflect", "os" & "syscall" are part-implemented, using an implementation of the nacl runtime.



| Name            | Passes tests in?      | Comment                           |
| --------------- | --------------------- | --------------------------------- |
| archive         | no code               |                                   |
| -- tar          | c++, js               | c#/java: issue relating to incorrect map handling in reflect |
| -- zip          | (c++), c#, java, js   | c++ test takes >15 mins           |
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
| -- aes          | c++, c#, java, js     |                                   |
| -- cipher       | c++, c#, java, js     |                                   |
| -- des          | c++, c#, java, js     |                                   |
| -- dsa          | c++, c#, java, js     |                                   |
| -- ecdsa        | c++, c#, java, js     |                                   | 
| -- elliptic     | c++, c#, java, js     |                                   |
| -- hmac         | c++, c#, java, js     |                                   |
| -- md5          | c++, c#, java, js     |                                   |
| -- rand         | c++, c#, java, js     |                                   |
| -- rc4          | c++, c#, java, js     |                                   |
| -- rsa          |                       | waiting for reflect.Call          |
| -- sha1         | c++, c#, java, js     |                                   |
| -- sha256       | c++, c#, java, js     |                                   |
| -- sha512       | c++, c#, java, js     |                                   |
| -- subtle       |                       | waiting for reflect.Call          |
| -- tls          |                       | panic: duplicate function name: crypto/tls.run$1 |
| -- x509         | (c#, js)              | modified tests (as Windows) js >30 mins; c++/java fail compile |
| -- -- pkix      | no tests              |                                   |
| database        | no code               |                                   |
| -- sql          |                       | panic: duplicate function name: database/sql.Query$1 |
| -- -- driver    | c++, c#, java, js     |                                   |
| debug           | no code               |                                   |
| -- dwarf        |                       | interface type assert failed      |
| -- elf          | c++, c#, js           | java: error TypeInfo code too large |
| -- gosym        | c++, c#, java, js     |                                   |
| -- macho        |                       | file_test.go 169: duplicate architecture |
| -- pe           | c++, c#, java, js     |                                   |
| -- plan9obj     | c++, c#, java, js     |                                   |
| encoding        | no tests              |                                   |
| -- ascii85      | c++, c#, java, js     |                                   |
| -- asn1         |                       | 2 errors, probably both UTF-8 encoding related |
| -- base32       | c++, c#, java, js     |                                   |
| -- base64       | c++, java, js         | c#: timeout                       |
| -- binary       |                       | reflect: unknown method using value from unexported field |
| -- csv          | c++, c#, java, js     |                                   |
| -- gob          |                       | fatal error: stack overflow       |
| -- hex          | c++, c#, java, js     |                                   |
| -- json         |                       | 2 errors related to seeing fields |
| -- pem          | c++, js               | c#/java: type cast exception Pointer/GOmap |
| -- xml          |                       | multiple errors, then crashes     |
| errors          | c++, c#, java, js     |                                   |
| expvar          |                       | Haxe try-catch exception after JSON unmarshall |
| flag            | c++, c#, java, js     | but no way to pass flags in yet   |
| fmt             | c++, js               | minor differences in type names, c#/java: error in reflect |
| go              | no code               |                                   |
| -- ast          |                       | multiple errors                   |
| -- build        |                       | $GOROOT/$GOPATH not set           |
| -- doc          |                       | os.PathError, probably test-data set-up issue |
| -- format       | c++, c#, java, js     |                                   |
| -- parser       |                       | os.PathError, probably test-data set-up issue |
| -- printer      |                       | os.PathError, probably test-data set-up issue |
| -- scanner      | c++, c#, java, js     |                                   |
| -- token        | c++, c#, java, js     |                                   |
| hash            | no tests              |                                   |
| -- adler32      | c++, c#, java, js     |                                   |
| -- crc32        | c++, c#, java, js     |                                   |
| -- crc64        | c++, c#, java, js     |                                   |
| -- fnv          | c++, c#, java, js     |                                   |
| html            | c++, c#, java, js     |                                   |
| -- template     |                       | waiting for reflect.Call          |
| image           | c++, c#, java, js     |                                   |
| -- color        | c++, c#, java, js     |                                   |
| -- -- palette   | no tests              |                                   |
| -- draw         | c++, c#, java, js     |                                   |
| -- gif          | c++, c#, java, js     |                                   |
| -- jpeg         | c++, c#, java, js     |                                   |
| -- png          |                       | concrete type assert failed       |
| index           | no code               |                                   |
| -- suffixarray  | c++, c#, java, js     |                                   |
| io              | c++, c#, java, js     |                                   |
| -- ioutil       | c++, c#, java, js     |                                   |
| log             |                       | multiple matching errors          |
| -- syslog       | no tests              |                                   |
| math            | c++, js               | c#/java: float32/int overflow issues |
| -- big          |                       | waiting for reflect.Call          |
| -- cmplx        | c++, c#, java, js     |                                   |
| -- rand         |                       | waiting for reflect.Method        |
| mime            | c++, c#, java, js     |                                   |
| -- multipart    |                       | hangs                             |
| net             |                       | hangs                             |
| -- http         |                       | M not declared by dummy package testing |
| -- -- cgi       | no                    | fork/exec not implemented         |
| -- -- cookiejar |                       | errors in TestUpdateAndDelete     |
| -- -- fcgi      | js                    | other targets fail or take too long to compile |
| -- -- httptest  |                       | hangs                             |
| -- -- httputil  |                       | hangs                             |
| -- -- internal  | c++, c#, java, js     |                                   |
| -- -- pprof     | no tests              |                                   |
| -- mail         | c++, c#, java, js     |                                   |
| -- rpc          |                       | hangs                             |
| -- -- jsonrpc   |                       | hangs                             |
| -- smtp         |                       | panic: syscall.stopTimer()        |
| -- textproto    | c++, c#, java, js     |                                   |
| -- url          | c++, c#, java, js     |                                   |
| os              | c++, c#, java, js     | passes modified tests (no system files to read) |
| -- exec         | -                     | tests fail, dummy testing T.Skip() not properly implemented |
| -- signal       | -                     | no tests (for nacl)               |
| -- user         | -                     | tests run with (correct) errors   |
| path            | c++, c#, java, js     |                                   |
| -- filepath     | c++, c#, java, js     |                                   |
| reflect         |                       | partially implemenented - 1st error: invalid function reference |
| regexp          | c++, c#, java, js     |                                   |
| -- syntax       | c++, c#, java, js     |                                   |
| runtime         | some                  | some general tests pass, NaN handled differently as a Map key |
| -- cgo          | -                     | unsupported                       |
| -- debug        | -                     | unsupported                       |
| -- pprof        | -                     | unsupported                       |
| -- race         | -                     | unsupported                       |
| sort            | c++, c#, java, js     |                                   |
| strconv         | c++, js               | c#/java: float32 issues           |
| strings         | c++, c#, java, js     |                                   |
| sync            |                       | hangs                             |
| -- atomic       | c++, c#, java, js     |                                   |
| syscall         |                       | partial implementation via nacl   |
| testing         |                       | dummy at present                  |
| -- iotest       |                       | dummy                             |
| -- quick        |                       | dummy                             |
| text            | no code               |                                   |
| -- scanner      | c++, c#, java, js     |                                   |
| -- tabwriter    | c++, c#, java, js     |                                   |
| -- template     |                       | hangs                             |
| -- -- parse     |                       | 2 errors related to integer 1e19  |
| time            |                       | duration error and waiting for reflect.Call |
| unicode         | c++, c#, java, js     |                                   |
| -- utf16        | c++, c#, java, js     |                                   |
| -- utf8         | c++, c#, java, js     |                                   |
| unsafe          | no tests              | pointer arithmetic unsupported, other functionalty should work |

(With thanks to [GopherJS](https://github.com/gopherjs/gopherjs/blob/master/doc/packages.md) for the layout above)
