// Package remctl provides access to the remctl client library.
// http://www.eyrie.org/~eagle/software/remctl/
//
// Copyright 2014 Thomas L. Kula <kula@tproa.net>
// All Rights Reserved
//
// Use of this source code is governed by a BSD-style license that can 
// be found in the LICENSE file.


/*
    Package remctl provides access to the remctl client library from
    http://www.eyrie.org/~eagle/software/remctl/

    There is a 'simple' interface which effectively implements the
    'remctl' and 'remctl_result_free' functions from the C api
    (in Go programs remctl_result_free is called for you). 
*/
package remctl

/*
#cgo pkg-config: libremctl
#include <remctl.h>
#include <stdlib.h>
#include <sys/uio.h>

// Cheerfully stolen from https://groups.google.com/forum/#!topic/golang-nuts/pQueMFdY0mk 
static char** makeCharArray( int size ) {
    // We need to hold 'size' character arrays, and 
    // null terminate the last entry for the remctl library

    char** a = calloc( sizeof(char*), size + 1 );
    a[size] = NULL;

    return a;
}

static void setArrayString( char **a, char *s, int n ) {
    a[n] = s;
}

static void freeCharArray( char **a, int size ) {
    int i;
    for( i = 0; i < size; i++ ) {
	free( a[i] );
    }
    free(a);
}


*/
import "C"

import (
    "errors"
    "fmt"
    "unsafe"
)

// The "Simple" interface

// RemctlResult is what is returned by the remctl 'simple' interface.
// Stdout and Stderr are the returned stdout and stderr, respectively,
// while Status is the return code.
type RemctlResult struct {
    Stdout  string
    Stderr  string
    Status  int
}

// Simple is the 'simple' interface.
//  If port is 0, the default remctl port is used
//  If principal is "", the default principal of "host/<hostname>" is used
//
// For more control of how the call is made, use the more 'complex'
// interface.
func Simple( host string, port uint16, principal string, command []string ) ( *RemctlResult, error ) {
    var res *C.struct_remctl_result

    host_c := C.CString( host )
    defer C.free( unsafe.Pointer( host_c ))


    var principal_c *_Ctype_char
    if principal != "" {
	principal_c = C.CString( principal )
	defer C.free( unsafe.Pointer( principal_c ))
    }

    command_len := len( command )
    command_c := C.makeCharArray( C.int( command_len ))
    defer C.freeCharArray( command_c, C.int( command_len ))
    for i, s := range command {
	C.setArrayString( command_c, C.CString( s ), C.int( i ))
    }

    res, err := C.remctl( host_c, C.ushort( port ), principal_c, command_c )

    if res == nil {
	return nil, err
    } else {
	defer C.remctl_result_free((*C.struct_remctl_result)(unsafe.Pointer( res )))
	result := &RemctlResult{}
	stdout_len := (C.int)(res.stdout_len)
	stderr_len := (C.int)(res.stderr_len)
	result.Stdout = C.GoStringN( res.stdout_buf, stdout_len )
	result.Stderr = C.GoStringN( res.stderr_buf, stderr_len )
	result.Status = (int)(res.status)
	if res.error != nil {
	    return nil, errors.New( C.GoString( res.error ))
	}

	return result, nil
    }
}

// The 'complex' interface

type Command []string

type Output struct {
    data *string
    stream int
}

type Error struct {
    data *string
    code int
}

type Done bool

type Status int

const (
    REMCTL_OUT_OUTPUT = iota
    REMCTL_OUT_STATUS
    REMCTL_OUT_ERROR
    REMCTL_OUT_DONE
)

// The following are from the remctl protocol specification
const (
    ERROR_NOCODE = iota		// 0 isn't used
    ERROR_INTERNAL		// Internal server failure
    ERROR_BAD_TOKEN		// Invalid format in token
    ERROR_UNKNOWN		// Unknown message type
    ERROR_BAD_COMMAND		// Invalid command format in token
    ERROR_UNKNOWN_COMMAND	// Command not defined on remctld server
    ERROR_ACCESS		// Access denied
    ERROR_TOOMANY_ARGS		// Argument count exceeds server limit
    ERROR_TOOMUCH_DATA		// Argument size exceeds server limit
    ERROR_UNEXPECTED_MESSAGE	// Message type not valid now
)

type remctl struct {
    ctx	    *C.struct_remctl
    Output chan Output
    Error chan Error
    Status chan Status
    Done chan Done  // We send true over this when REMCTL_OUT_DONE is returned
    open bool
}

func (r *remctl) get_error() ( error ) {
    return errors.New( C.GoString( C.remctl_error( r.ctx )))
}

func (r *remctl) assert_init() ( error ) {
    if r.ctx == nil {
	return errors.New( "Connection not inited" )
    }

    return nil
}

func (r *remctl) isopen() ( bool ) {
    return r.open
}

// All of the remctl calls have the convention that if the principal is
// NULL, use the default. we use the convention that if the principal is
// an empty string, we use the default. Translate.
func get_principal ( principal string ) (*_Ctype_char) {
    var principal_c *_Ctype_char
    if principal != "" {
	principal_c = C.CString( principal )
    }

    return principal_c
}

// Free a principal, but only if it's not NULL
func free_principal ( principal_c *_Ctype_char ) {
    if principal_c != nil {
	C.free( unsafe.Pointer( principal_c ))
    }
}

// Create a new remctl connection
func New() (*remctl, error ) {
    r := &remctl{}
    r.Output = make( chan Output )
    r.Error = make( chan Error )
    r.Done = make( chan Done )
    r.Status = make( chan Status )

    ctx, err := C.remctl_new()
    r.ctx = ctx
    r.open = false
    return r, err
}

// Setccache allows you to define the credentials cache the underlying
// GSSAPI library will use to make the connection.
//
// Your underlying GSSAPI library may not tolerate that, so be prepared
// for errors
func (r *remctl) Setccache( ccache string ) ( error ) {
    ccache_c := C.CString( ccache )
    defer C.free( unsafe.Pointer( ccache_c ))

    if set := C.remctl_set_ccache( r.ctx, ccache_c ); set != 1 {
	return r.get_error()
    }

    return nil
}

// Close() the remctl connection after you are done with it
// 
// After calling, the struct is useless and should be deleted
func (r *remctl) Close() (error) {
    if open := r.assert_init(); open != nil {
	return open
    }

    C.remctl_close( r.ctx ) // The remctl library frees the memory pointed to by r.ctx
    r.ctx = nil
    return nil
}

// Open() a connection
// 
// Open a connection to `host' on port `port'. If port is 0, use the default
// remctl port. You may specify a principal to use in `principal', if it is
// a blank string the remctl will use the default principal. 
func (r *remctl) Open( host string, port uint16, principal string ) ( error ) {
    if r.isopen() { 
	return errors.New( "Already open" )
    }

    host_c := C.CString( host )
    defer C.free( unsafe.Pointer( host_c ))
    port_c := C.ushort( port )

    principal_c := get_principal( principal )
    if principal != "" {
	// If principal is empty, principal_c is NULL, don't free
	defer C.free( unsafe.Pointer( principal_c ))
    }

    if opened := C.remctl_open( r.ctx, host_c, port_c, principal_c ); opened != 1 {
	return r.get_error()
    }

    return nil
}

// Execute() a command
// 
// Executes a command whose arguments are a slice of strings. This starts a
// goroutine which will send Output over the remctl.Output channel.  It will
// continue to do this until there is no more output, after which it will send
// Status over the remctl.Status channel. If at any point there is an error
// message, it will send an Error over the remctl.Error channel. If the remctl
// server things the connection is done, `true' will be sent over the Done
// channel. In any case, after a single Status, Error or Done has been
// returned, the command is done and the remctl connection is ready for the
// next Execute() or Close()
//
// If this function returns an error, nothing except Close() is safe
func (r *remctl) Execute( cmd Command ) (error) {


    if !r.isopen() {
	return errors.New( "Not open" )
    }

    // Idea cheerfully stolen from go_fuse's syscall_linux.go
    iov := make([]_Ctype_struct_iovec, len(cmd))
    for n, v := range cmd {
	cmd_c := C.CString(v)
	defer C.free(unsafe.Pointer(cmd_c))
	iov[n].iov_base = unsafe.Pointer(cmd_c)
	iov[n].iov_len = C.size_t(len(v))
    }

    if sent := C.remctl_commandv( r.ctx, &iov[0], C.size_t(len(cmd))); sent == 1 {
	// It failed, return an error
	return r.get_error()
    }

    // Now we enter a goroutine that pulls various bits of data out of
    // the remctl connection and sends it down channels until we're
    // done
    go func(r *remctl) {
	var output *C.struct_remctl_output
	for {
	    output = C.remctl_output(r.ctx)
	    if output == nil {
		error_msg := Error{}
		err_txt := fmt.Sprintf( "%s", r.get_error())
		error_msg.data = &err_txt
		error_msg.code = ERROR_NOCODE   // We fake this here
		r.Error <- error_msg
		// And we're done
		return
	    }

	    switch output_type := C.int( output._type ); output_type {
	    case REMCTL_OUT_OUTPUT:
		output_msg := Output{}
		data := C.GoStringN( output.data, C.int(output.length))
		output_msg.data = &data
		output_msg.stream = int( output.stream )
		r.Output <- output_msg
	    case REMCTL_OUT_STATUS:
		r.Status <- Status( output.status )
		return
	    case REMCTL_OUT_ERROR:
		error_msg := Error{}
		data := C.GoStringN( output.data, C.int(output.length ))
		error_msg.data = &data
		error_msg.code = int( output.error )
		r.Error <- error_msg
		return
	    case REMCTL_OUT_DONE:
		r.Done <- true
		return
	    }
	}
    }(r) // End of goroutine

    return nil
}
