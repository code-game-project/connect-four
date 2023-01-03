# Connect 4
![CodeGame Version](https://img.shields.io/badge/CodeGame-v0.7-orange)
![CGE Version](https://img.shields.io/badge/CGE-v0.4-green)

Drop colored tokens into a grid. You win when you manage to form a horizontal, vertical or diagonal line of four tokens.

## Known instances

- `games.code-game.org/connect-four`

## Usage

```sh
# Run on default port 8080
connect-four

# Specify a custom port
connect-four --port=5000

# Specify a custom port through an environment variable
CG_PORT=5000 connect-four
```

### Running with Docker

Prerequisites:
- [Docker](https://docker.com/)

```sh
# Download image
docker pull codegameproject/connect-four:0.1

# Run container
docker run -d --restart on-failure -p <port-on-host-machine>:8080 --name connect-four codegameproject/connect-four:0.1
```

## Event Flow

1. You receive a `start` event when a second player joins, which includes your color ('yellow' or 'red').
2. You regularly receive a `grid` event, which includes the current state of the grid.
3. You receive a `turn` event, which includes the next sign to be placed.
4. When it is your turn you can send a `drop_token` event with the column in which you want to drop your token.
5. When the game is complete you will receive a `game_over` event, which includes which player wins and which cells form the winning line. Otherwise go to 3.
6. You will receive an `invalid_action` event, if you try to do something that's not allowed like dropping a piece when it is not your turn.

## Building

### Prerequisites

- [Go](https://go.dev) 1.19+

```sh
git clone https://github.com/code-game-project/connect-four.git
cd connect-four
codegame build
```

## License

Copyright (C) 2023 Julian Hofmann

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
