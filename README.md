# go-remctl

A Go interface to the remctl client library, version 1.0

Provides an interface to the [remctl] client library, a 'client/server
application that supports remote execution of specific commands, using Kerberos
GSS-API for authentication and confidentiality.'. Provides both a _simple_
interface and a _complex_ interface which allows more control over the remctl
interaction.

Requires the remctl C client libraries, and is designed to work with remctl
v2 protocol compliant servers.

[remctl]: http://www.eyrie.org/~eagle/software/remctl/

## Example usage

See the *example* directory for usage. 

## License

go-remctl is under a two-clause BSD license. See the [LICENSE][license] file
for details.

[license]: https://github.com/kula/go-remctl/blob/master/LICENSE
