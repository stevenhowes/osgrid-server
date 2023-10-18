# osgrid-server - _Ordnance Survey Grid Ref Converter_

This project implements a simple REST server to allow conversion between Ordnance Survey
grid references and latitude / longitude. This fork runs without docker and includes
a makefile implementing install, run, build, vet, lint, fmt and clean. As this is part
of an intended set of SAR tools, installation is to /opt/sartools/ and will create the
user 'sartools' if it does not exist.

## Pre-requisites
- go
- make
- golint (optional)

## Usage

To use this stand-alone (non Docker variant) on http://localhost:9090 just run:

    make install
    service osgrid-server

A grid reference may then be converted to a lat / lon:

    $ curl http://localhost:9090/v4/gridref/SU37241550
    {"osGridRef":"SU37241550","easting":437240,"northing":115500,"lat":50.93774069083967,"lon":-1.4713807843405298}

Or a lat/lon converted to a grid reference:

    $ curl http://localhost:9090/v4/latlon/50.9378,-1.4713
    {"osGridRef":"SU37241550","easting":437246,"northing":115507,"lat":50.9378,"lon":-1.4713}

