#!/usr/local/bin/gentee

// This script builds releases of Run Buttons application
// It uses Gentee programming language - https://github.com/gentee/gentee

struct osArch {
    str os
    str arch
}

const : RELEASE = `/home/ak/releases/run-buttons`

run {
    ChDir(`/home/ak/go/github.com/gentee/run-buttons`)
    arr.osArch list = {
        {os: `linux`, arch: `amd64`}
        {os: `linux`, arch: `386`}
        {os: `windows`, arch: `amd64`}
        {os: `windows`, arch: `386`}
        {os: `darwin`, arch: `amd64`}
    }
    for item in list {
        ChDir(`/home/ak/go/github.com/gentee/run-buttons`)
        str appname = `run-buttons` + ?(item.os == `windows`, `.exe`, `` )
        $GOOS = item.os
        $GOARCH = item.arch
        $ go build -o %{RELEASE}/%{appname}
        str src = ReadFile(`run-buttons.go`)
        CreateDir(RELEASE)
        ChDir(RELEASE)
        arr.arr.str ver &= FindRegExp(src, `Version\s+=\s*"([\d\.]+)"`)
        $ zip run-buttons-%{ver[0][1]}-%{item.os}-%{item.arch}.zip -m %{appname}
    }
}