# VBScript Tutorial

This tutorial describes how to use AI Mastering API through AI Mastering CLI client from VBScript. 

## Glossary

### AI Mastering

An automated mastering service.
https://aimastering.com/

### AI Mastering API

The web API of AI Mastering.

### AI Mastering CLI Client

A command line client of AI Mastering API.
This makes easy to use the API.

## How to

### Install AI Mastering CLI Client

Please download

[AI Mastering CLI client for windows](https://github.com/ai-mastering/aimastering-tools/releases/download/v1.0.1/aimastering-windows-386.exe)

### Get AI Mastering API access token

The access token is needed to access AI Mastering API.
So, please get the access token from https://aimastering.com/app/developer.
Please note that this token should be secret.

### Run AI Mastering CLI Tool from VBScript

VBScript Example

[master.vbs](/examples/vbscript/master.vbs)

Preliminary

- Move AI Mastering CLI client executable to C:\aimastering-windows-386.exe
- Place input file to C:\input.wav

And execute following

```bash
SET AIMASTERING_ACCESS_TOKEN=[AI Mastering API access token]
wscript examples/vbscript/master.vbs
```

If commands succeeded, the mastered audio file is created in C:\output.wav. 

### More options

Please see the output of following commands.

```bash
aimastering-windows-386.exe --help 
aimastering-windows-386.exe master --help 
```
