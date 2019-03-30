# AI Mastering tools

This repository includes

- Command line tools to execute automated mastering using [AI Mastering](https://aimastering.com) API
- GUI front-end (Coming Soon)
- Source code of above

This repository may be helpful as to how to use the AI Mastering API.

## Install

### Windows

Please download executable binary.

https://github.com/ai-mastering/aimastering-tools/releases/download/v1.0.1/aimastering-windows-386.exe

### Mac

Please execute following commands.

```bash
sudo curl -L "https://github.com/ai-mastering/aimastering-tools/releases/download/v1.0.1/aimastering-darwin-386" -o /usr/local/bin/aimastering
sudo chmod +x /usr/local/bin/aimastering
```

### Linux

Please execute following commands.

```bash
sudo curl -L "https://github.com/ai-mastering/aimastering-tools/releases/download/v1.0.1/aimastering-linux-386" -o /usr/local/bin/aimastering
sudo chmod +x /usr/local/bin/aimastering
```

### Bash completion

Please add following command to ~/.bash_profile

```bash
eval "$(aimastering autocomplete --shell bash)"
```

### Zsh completion (not tested)

Please add following command to ~/.zshenv

```bash
eval "$(aimastering autocomplete --shell zsh)"
```

## Command line tool usage

### Auth

Please set AIMASTERING_ACCESS_TOKEN env var.
access token can be retrieved https://aimastering.com/app/developer.

```bash
export AIMASTERING_ACCESS_TOKEN=xxx
```

Access token can also be passed by --access-token options.

### Basic

#### Pass access token by env var

```bash
export AIMASTERING_ACCESS_TOKEN=xxx
aimastering master --input /path/to/input.wav --output /path/to/output.wav
```

#### Pass access token by argument

```bash
aimastering master --input /path/to/input.wav --output /path/to/output.wav --access-token=xxx
```

### Options

#### Target Loudness

Target loudness -6dB

```bash
aimastering master --input input.wav --output output.wav --target-loudness -6
```

#### Other options

Please see

```bash
aimastering --help
aimastering master --help
```

## GUI usage (Coming Soon)


## Requirements

No dependencies

## Notes

This tool is an auxiliary tool of [AI Mastering](https://aimastering.com).
We do not guarantee the maintenance of this tool.

## LICENSE

MIT
