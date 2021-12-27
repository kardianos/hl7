#!/bin/bash
go run ./hl7fetch/*.go -pkgdir v21 -root ./genjson -version 2.1
go run ./hl7fetch/*.go -pkgdir v231 -root ./genjson -version 2.3.1
go run ./hl7fetch/*.go -pkgdir v25 -root ./genjson -version 2.5
go run ./hl7fetch/*.go -pkgdir v251 -root ./genjson -version 2.5.1
go run ./hl7fetch/*.go -pkgdir v28 -root ./genjson -version 2.8
