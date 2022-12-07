#!/bin/bash
go run ./hl7fetch/*.go -pkgdir h210 -root ./genjson -version 2.1
go run ./hl7fetch/*.go -pkgdir h220 -root ./genjson -version 2.2
go run ./hl7fetch/*.go -pkgdir h231 -root ./genjson -version 2.3.1
go run ./hl7fetch/*.go -pkgdir h240 -root ./genjson -version 2.4
go run ./hl7fetch/*.go -pkgdir h250 -root ./genjson -version 2.5
go run ./hl7fetch/*.go -pkgdir h251 -root ./genjson -version 2.5.1
go run ./hl7fetch/*.go -pkgdir h270 -root ./genjson -version 2.7
go run ./hl7fetch/*.go -pkgdir h271 -root ./genjson -version 2.7.1
go run ./hl7fetch/*.go -pkgdir h280 -root ./genjson -version 2.8
