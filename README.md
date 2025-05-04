# yabc

> Yet Another Bluesky CLI

A command-line interface for interacting with the Bluesky social network.

## Installation

### Using Go

```bash
go install github.com/alexisbcz/yabc@latest
```

### Using Goblin.run

```bash
curl -sf http://goblin.run/github.com/alexisbcz/yabc | sh
```

## Configuration

Before using yabc, you need to set up your Bluesky credentials as environment variables:

```bash
export BLUESKY_IDENTIFIER="your-handle.bsky.social"
export BLUESKY_PASSWORD="your-app-password"
```

It's recommended to add these to your `.bashrc`, `.zshrc`, or appropriate shell configuration file.

## Usage

yabc provides various commands for interacting with Bluesky:

### Posts

Create a new post:

```bash
yabc posts create --text "Hello world!" --hashtags coding,golang
```

Create a post with an image (when supported):

```bash
yabc posts create --text "Check out this photo" --image path/to/image.jpg
```

### More Commands

For a full list of available commands:

```bash
yabc help
```

Get help for a specific command:

```bash
yabc posts help
```

## Documentation

For complete documentation, run:

```bash
yabc help
```

This will show all available commands and their usage details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create a new Pull Request

## License

[MIT](./LICENSE)

## Author

Alexis Bouchez (`alexbcz@proton.me`) — [alexisbouchez.com](https://alexisbouchez.com)
