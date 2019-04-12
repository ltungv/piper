# VGU Rocobon 2019 Data Broadcasting Server

Simple broadcasting server using WebSocket protocol

With support for running script and broadcasting its output

# How to install

```bash
go get -u github.com/letung3105/piper/cmd/piper
```

## For development

Please read the `Makefile` for more information

Build
```bash
go get -u github.com/letung3105/piper
cd $GOPATH/src/github.com/letung3105/piper
make build
```

Install
```bash
go get -u github.com/letung3105/piper
cd $GOPATH/src/github.com/letung3105/piper
make install
```

# Usage

```bash
Usage of piper:
  -b string
        name of interpreter (default "python")
  -p string
        port use (default "8000")
  -s string
        name of script
```

## Example
```bash
piper -b python3 -s ./scripts/mock.py -p 8000
```
