<p align="center">
    <img width="400" src="assets/cover.png" />
</p>

👋 Hello! This is an app for playing a game of poker! It was created as an exercise in sniffing network traffic in an unsecured network but it turned out to be a full-fletched application for having some fun!

## ❤️ Getting started

Clone the repository:

```
git clone https://github.com/TypicalAM/gopoker
```

Run the backend:

```
cd backend
go run cmd/gopoker/gopoker.go
```

and the frontend:

```
cd frontend
npm start
```

You can now explore the page at `localhost:3000`

## Technologies

This is my first adventure with front-end so the UI part isn't that impressive, the tech used is the following:
- ReactJS
- TailwindCSS
- Typescript

The backend is written in go with:
- Authorization and middlewares
- Gin
- Gorm

## 📸 Here is a demo

Here is some basic usage:

<p align="center">
    <img src="assets/basic game.png" />
</p>

<p align="center">
    <img src="assets/basic game 2.png" />
</p>

