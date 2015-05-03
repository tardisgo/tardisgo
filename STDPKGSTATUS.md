Standard Package Status
-----------------------

The standard packages that [pass their tests (see end of file for automated summary)](https://github.com/tardisgo/tardisgo/blob/master/goroot/haxe/go1.4/src/tgotests.log) in js, c#, c++, or java are shown below. 

The "testing" package is emulated in an ugly and part-working way, all tests are run in Short mode. Packages "reflect", "os" & "syscall" are part-implemented, using an implementation of the nacl runtime.

The [standard library tests](https://github.com/tardisgo/tardisgo/blob/master/goroot/haxe/go1.4/src/tgotests.go) take more than 2 hours to run on an 8 core Mac.

Some tests marked "*" below use testdata in the pseudo file system, passed in via a local "tgotestfs.zip".

Tests bracketed by "[]" work, but currently take too long, so are excluded from the automated tests.

| Name            | Passes in? (*=testfs) | Comment                           |
| --------------- | --------------------- | --------------------------------- |
| archive         | no code               |                                   |
| -- tar          | c++, js, java       * | c#: header mis-match, ModTime     |
| -- zip          | [c++] c#, java, js  * | c++: tests take >30mins           |
| bufio           | c++, c#, java, js     |                                   |
| builtin         | no tests              | all built-in functions are implemented |
| bytes           | c++, c#, java, js     |                                   |
| compress        | no code               |                                   |
| -- bzip2        | c++, c#, java, js   * |                                   |
| -- flate        | c++, c#, java, js   * |                                   |
| -- gzip         | c++, c#, java, js   * |                                   |
| -- lzw          | c++, c#, java, js   * |                                   |
| -- zlib         | c++, c#, java, js   * |                                   |
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
| -- rsa          |                     * | try-catch exception               |
| -- sha1         | c++, c#, java, js     |                                   |
| -- sha256       | c++, c#, java, js     |                                   |
| -- sha512       | c++, c#, java, js     |                                   |
| -- subtle       | c++, c#, java, js     |                                   |
| -- tls          |                     * | panic: duplicate function name: crypto/tls.run$1 |
| -- x509         | [c++, c#, java, js]   | mod tests (as Windows) but slow js>30m |
| -- -- pkix      | no tests              |                                   |
| database        | no code               |                                   |
| -- sql          |                       | panic: duplicate function name: database/sql.Query$1 |
| -- -- driver    | c++, c#, java, js     |                                   |
| debug           | no code               |                                   |
| -- dwarf        | c++, c#, java, js   * |                                   |
| -- elf          | c#, java, js        * | c++: clang seg fault              |
| -- gosym        | c++, c#, java, js     |                                   |
| -- macho        | c++, c#, java, js   * |                                   |
| -- pe           | c++, c#, java, js   * |                                   |
| -- plan9obj     | c++, c#, java, js   * |                                   |
| encoding        | no tests              |                                   |
| -- ascii85      | c++, c#, java, js     |                                   |
| -- asn1         | c++, c#, java, js     |                                   |
| -- base32       | c++, c#, java, js     |                                   |
| -- base64       | c++, c#, java, js     |                                   |
| -- binary       | c++, c#, java, js     |                                   |
| -- csv          | c++, c#, java, js     |                                   |
| -- gob          |                       | reflect.mapassign() null Haxe map |
| -- hex          | c++, c#, java, js     |                                   |
| -- json         |                     * | field name not found, then fatal  |
| -- pem          | c++, c#, java, js     |                                   |
| -- xml          |                       | multiple errors, then fatal       |
| errors          | c++, c#, java, js     |                                   |
| expvar          | js, cs                | c++: clang seg fault, java: haxe compiler hangs |
| flag            | c++, c#, java, js     | but no way to pass flags in yet   |
| fmt             | c++, c#, java, js     | minor differences in type names   |
| go              | no code               |                                   |
| -- ast          | c++, c#, java, js     | minor changes to testdata whitespace and paths |
| -- build        |                     * | $GOROOT/$GOPATH not set           |
| -- doc          | c++, java, js       * | cs: matching issue                |
| -- format       | c++, c#, java, js   * |                                   |
| -- parser       | c++, c#, java, js   * |                                   |
| -- printer      | c++, c#, java, js   * |                                   |
| -- scanner      | c++, c#, java, js     |                                   |
| -- token        | c++, c#, java, js     |                                   |
| hash            | no tests              |                                   |
| -- adler32      | c++, c#, java, js     |                                   |
| -- crc32        | c++, c#, java, js     |                                   |
| -- crc64        | c++, c#, java, js     |                                   |
| -- fnv          | c++, c#, java, js     |                                   |
| html            | c++, c#, java, js     |                                   |
| -- template     |                       | a number of errors and missing parts of reflect |
| image           | c++, c#, java, js   * |                                   |
| -- color        | c++, c#, java, js     |                                   |
| -- -- palette   | no tests              |                                   |
| -- draw         | c++, c#, java, js   * |                                   |
| -- gif          | c++, c#, java, js   * |                                   |
| -- jpeg         | c++, c#, java, js   * |                                   |
| -- png          |                     * | concrete type assert failed       |
| index           | no code               |                                   |
| -- suffixarray  | c++, c#, java, js     |                                   |
| io              | c++, c#, java, js     |                                   |
| -- ioutil       | c++, c#, java, js   * |                                   |
| log             | c++, c#, java, js     |                                   |
| -- syslog       | no tests              |                                   |
| math            | c++, js               | c#/java: float32/int overflow issues |
| -- big          |                       | fmt error then fails on gob encode |
| -- cmplx        | c++, c#, java, js     |                                   |
| -- rand         |                       | waiting for reflect.makeMethodValue |
| mime            | c++, c#, java, js   * |                                   |
| -- multipart    |                     * | hangs                             |
| net             |                       | hangs                             |
| -- http         |                       | M not declared by dummy package testing |
| -- -- cgi       | -                     | fork/exec not implemented         |
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
| os              | c++, c#, java, js   * | passes modified tests (no system files to read) |
| -- exec         | -                     | tests fail, dummy testing T.Skip() not properly implemented |
| -- signal       | -                     | no tests (for nacl)               |
| -- user         | -                     | tests run with (correct) errors   |
| path            | c++, c#, java, js     |                                   |
| -- filepath     | c++, c#, java, js   * |                                   |
| reflect         |                       | partially implemented             |
| regexp          | c++, c#, java, js   * |                                   |
| -- syntax       | c++, c#, java, js     |                                   |
| runtime         | (c++, c#, java, js)   | only a sub-set of tests pass, NaN Map key handled differently |
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
| -- template     |                     * | no value errors and awaiting reflect.makeMethodValue |
| -- -- parse     | c++, c#, java, js     |                                   |
| time            |                     * | nil pointer dereference           |
| unicode         | c++, c#, java, js     |                                   |
| -- utf16        | c++, c#, java, js     |                                   |
| -- utf8         | c++, c#, java, js     |                                   |
| unsafe          | no tests              | pointer arithmetic unsupported, other functionalty should work |

(With thanks to [GopherJS](https://github.com/gopherjs/gopherjs/blob/master/doc/packages.md) for the layout above)
