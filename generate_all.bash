#!/bin/bash
go run ./hl7fetch/*.go -pkgdir h21 -root ./genjson -version 2.1
go run ./hl7fetch/*.go -pkgdir h231 -root ./genjson -version 2.3.1
go run ./hl7fetch/*.go -pkgdir h25 -root ./genjson -version 2.5
go run ./hl7fetch/*.go -pkgdir h251 -root ./genjson -version 2.5.1
go run ./hl7fetch/*.go -pkgdir h28 -root ./genjson -version 2.8
