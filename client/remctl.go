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

// Cheerfully stolen from https://groups.google.com/forum/#!topic/golang-nuts/pQueMFdY0mk 
static char** makeCharArray( int size ) {
    return calloc( sizeof(char*), size );
}

static void setArrayString( char **a, char *s, int n ) {
    a[n] = s;
}

static void setArrayNull( char **a, int n ) {
    a[n] = NULL;
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

// Remctl is the 'simple' interface.
//  If port is 0, the default remctl port is used
//  If principal is "", the default principal of "host/<hostname>" is used
func Remctl( host string, port uint16, principal string, command []string ) ( *RemctlResult, error ) {
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
    C.setArrayNull( command_c, C.int( command_len ))

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


