// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package internal

/*
These internal packages abstract the following from the application (using
dependency injection):

 - error handling (for non-control flow)
 - trace support (non-localized output)

The above abstractions enable application code to not have to sprinkle
if (err != nil) blocks (except when the application wants to affect application
flow based on err)

Do and Do Not:
 - Do verify parameter values and panic if these internal functions would be unable
   to succeed, to catch coding errors (do not panic for user input errors)
 - Do not output (except for in the `internal/output` package). Do use the injected
    trace method to output low level debugging information
 - Do not return error if client is not going use the error for control flow, call the
   injected checkErr instead, which will probably end up calling cobra.checkErr and exit:
     e.g. Do not sprinkle application (non-helper) code with:
       err, _ := fmt.printf("Hope this works")
       if (err != nil) {
         panic("How unlikely")
       }
     Do use the injected checkErr callback and let the application decide what to do
       err, _ := printf("Hope this works)
       checkErr(err)
 - Do not have an internal package take a dependency on another internal package
   unless they are building on each other, instead inject the needed capability in the
   internal.initiaize()
     e.g. Do not have the config package take a dependency on the secret package, instead
          inject the methods encrypt/decrypt to config in its initialize method, do not:

       package config

       import (
         "github.com/microsoft/go-sqlcmd/cmd/internal/secret"
       )

     Do instead:

       package config

       var encryptCallback func(plainText string) (cipherText string)
       var decryptCallback func(cipherText string) (secret string)

       func Initialize(
       encryptHandler func(plainText string) (cipherText string),
       decryptHandler func(cipherText string) (secret string),
*/
