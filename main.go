package main

import (
	//"gorm.io/driver/sqlite"

	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Developer struct{
	gorm.Model
	Name string
	Username string
	Password string
	Games []Game`gorm:"foreignKey:DeveloperRefer"`
}
type Genre struct{
	gorm.Model
	Name string
}
type Game struct{
	gorm.Model
	Name string
	CurrentPrice float32
	DeveloperRefer uint
	Genres []Genre `gorm:"many2many:game_genres;"`
}
type User struct{
	gorm.Model
	Username string
	Password string
	Games []Game `gorm:"many2many:user_games;"`
}
type ActiveUser struct{
  ID        uint
  UserID uint
  Username string
  Role string
}


func main() {
    dsn := "host=localhost user=egelen password=12345 dbname=postgres port=5432 sslmode=disable TimeZone=Asia/Shanghai"
     db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
   if err != nil {
     panic("failed to connect database")
   }

  // Migrate the schema
  // db.AutoMigrate(&ActiveUser{})
  // db.Create(&ActiveUser{ID: 1,Username: "null",Role: "null",UserID: 0})

  // Create
  // db.Create(&Product{Code: "D42", Price: 100})

  // // Read
  // var product Product
  // db.First(&product, 1) // find product with integer primary key
  // db.First(&product, "code = ?", "D42") // find product with code D42

  // // Update - update product's price to 200
  // db.Model(&product).Update("Price", 200)
  // // Update - update multiple fields
  // db.Model(&product).Updates(Product{Price: 200, Code: "F42"}) // non-zero fields
  // db.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})

  // // Delete - delete product
  // db.Delete(&product, 1)


  app := &cli.App{
    Commands: []*cli.Command{
          {
            Name:"migrate",
          Aliases: []string{"t"},
          Action: func(ctx *cli.Context) error {
              db.AutoMigrate(&ActiveUser{},&Genre{},&Game{},&Developer{},&User{})
            
            return nil
          },
        },
        //USER
        {
            Name:    "user",
            Aliases: []string{"t"},
            Usage:   "options for task templates",
            Subcommands: []*cli.Command{
              {
                Name:"buy",
                Usage: "<GameID>",
                Action: func(ctx *cli.Context) error {
                  var activeUser ActiveUser
                  db.Where("ID=?",1).First(&activeUser)
                  if activeUser.Role != "user" {
                    color.Yellow("You need to be signed in as a user")
                    return nil
                  }
                  gameID := ctx.Args().First()
                  var user User
                  db.Where("ID=?",activeUser.UserID).First(&user)
                  var game Game
                  db.Where("ID=?",gameID).Find(&game)
                  user.Games = append(user.Games, game)
                  db.Save(&user)
                  return nil
                },
              },
              {
                Name: "mygames",
                Usage: "Lists all the games you have",
                Action: func(ctx *cli.Context) error {
                  var activeUser ActiveUser  
                  db.Where("ID=?",1).First(&activeUser)
                  if activeUser.Role == "dev" {
                    color.Yellow("Cannot list games as developer. Use developer list to see your own games")
                    return nil
                  }
                  var user User
                  db.Where("ID=?",activeUser.UserID).Preload("Games").First(&user)
                  fmt.Println("ID  GAME")
                  for _,game := range user.Games{
                    fmt.Println(game.ID , "   ", game.Name , " ",)
                  }
                    return nil
  
                },
              },
              {
                Name:    "signup",
                Aliases: []string{"a"},
                Usage:   "<username> <password>",
                Action: func(cCtx *cli.Context) error {
                  var activeUser ActiveUser
                  db.Where("ID = ?", 1).First(&activeUser)
                  if activeUser.UserID != 0 {
                    color.Yellow("Can't signup while logged In!")
                    return nil
                  }
    
                  username:= cCtx.Args().First()
                  password := cCtx.Args().Get(1)
                  if len(username) < 5{
                    color.Red("username length must be longer than 5 char")
                    return nil
                  }
                  if len(password) < 8 {
                    color.Red("password length must be longer than 8 char")
                    return nil
                  }
                    result:= db.Create(&User{Username: cCtx.Args().First(),Password: cCtx.Args().Get(1)})
                    if int(result.RowsAffected) == 1 {
                    color.Green("User Created Successfully!")
                    return nil
                    }
                    color.Red("Something Went Wrong!")
                    return nil
                },
            },
            {
              Name:    "login",
              Aliases: []string{"l"},
              Usage:   "<username> <password>",
              Action: func(cCtx *cli.Context) error {
                var activeUser ActiveUser
                db.Where("ID = ?", 1).First(&activeUser)
                if activeUser.UserID != 0 {
                  color.Yellow("You Are Already Logged In!")
                  return nil
                }
                user := User{}
                username := cCtx.Args().First()
                password := cCtx.Args().Get(1)
                result:= db.Where("username = ?", username).First(&user)
                if result.Error != nil{
                  color.Red("Invalid Username")
                  return nil
                }
                if user.Password == password{
                  activeUser.UserID = user.ID
                  activeUser.Role = "user"
                  activeUser.Username =user.Username
                  db.Save(&activeUser)
                  color.Green("Successfull Login.. Welcome "+user.Username+"!")
                  return nil;
                }
                color.Red("Invalid Password")
                  return nil
              },
          },
            },
        },
        //DEVELOPER
        {
          Name: "developer",
          Aliases: []string{"d"},
          Usage: "developer commands",
          Subcommands: []*cli.Command{
            {
              Name: "newgame",
              Aliases: []string{"ng"},
              Usage: "<Name> <Price> <GenreID> <GenreID2> Create a new game",
              Action: func(ctx *cli.Context) error {
                var activeUser ActiveUser
                db.Where("ID=?",1).First(&activeUser)
                if activeUser.Role == "user" {
                  color.Yellow("You can not create a game as user")
                  return nil
                }
                gameName := ctx.Args().First()
                gamePrice,err := strconv.ParseFloat(ctx.Args().Get(1), 32)
                if err != nil{
                  color.Red("Mismatched Price Type")
                }
                genreID,err := strconv.ParseUint(ctx.Args().Get(2), 10, 64)
                if err != nil{
                  color.Red("Mismatched Genre Type")
                }

                genreID2,err := strconv.ParseUint(ctx.Args().Get(3), 10, 64)
                if err != nil{
                  color.Red("Mismatched Genre Type")
                }

                devID := activeUser.UserID

                var genreRef Genre
                err = db.Where("ID=?",genreID).First(&genreRef).Error
                if err != nil{
                  log.Printf(err.Error())
                }
                var genreRef2 Genre
                err = db.Where("ID=?",genreID2).First(&genreRef2).Error
                if err != nil{
                  log.Printf(err.Error())
                }
                var devRef Developer
                err = db.Where("ID=?",devID).First(&devRef).Error
                if err != nil{
                  log.Printf(err.Error())
                }
                game := Game{Name: gameName,CurrentPrice: (float32(gamePrice)),DeveloperRefer: devID, Genres: []Genre{genreRef}}
                db.Create(&game)               
                return nil
              },
            },
            {
              Name:"newgenre",
              Usage: "<Name>",
              Action: func(ctx *cli.Context) error {
                var activeUser ActiveUser
                db.Where("ID=?",1).Find(&activeUser)
                if activeUser.Role != "dev" {
                  color.Yellow("You need to be signed in as a developer to add genre")
                }

                genre := Genre{Name: ctx.Args().First()}
                db.Save(&genre)
                color.Green("New genre has succsessfully created!")
                return nil
              },
            },
            {
              Name:"updategame",
              Usage:"<gameID> <Name> | <gameID> <Name> <Price>",
              Action: func(ctx *cli.Context) error {
                var activeUser ActiveUser
                db.Where("ID=?",1).Find(&activeUser)
                if activeUser.Role != "dev" {
                  color.Yellow("You need to be signed in as a developer to update game")
                }

                var game Game
                db.Where("ID=?",ctx.Args().First()).First(&game)
                if game.DeveloperRefer != activeUser.UserID{
                  color.Yellow("You cannot change other's games")
                }
                game.Name = ctx.Args().Get(1)

                if ctx.Args().Len() > 2{
                  float , _ := strconv.ParseFloat(ctx.Args().Get(2),32)
                  game.CurrentPrice = float32(float)
                }
                db.Save(&game)
                color.Green("Game has successfully Updated!")
                
                return nil
              },
            },
            {
              Name: "mygames",
              Usage: "Lists all your games",
              Aliases: []string{"li"},
              Action: func(cCtx *cli.Context) error {
                var activeUser ActiveUser
                db.Where("ID = ?", 1).First(&activeUser)
                if activeUser.Role != "dev" {
                  color.Yellow("You need to be developer!")
                  return nil
                }

                var developer Developer
                db.Where("ID=?",activeUser.UserID).Preload("Games").Find(&developer)

                fmt.Println("ID    PRICE    GAME          PublishDate")
                for _, game := range developer.Games{
                  fmt.Println(game.ID,"     ",game.CurrentPrice,"     " ,game.Name,"   ",game.CreatedAt)
                }
                
                  return nil
              },
            },{
              Name: "login",
              Usage: "<username> <password>",
              Aliases: []string{"l"},
              Action: func(cCtx *cli.Context) error {
                var activeUser ActiveUser
                db.Where("ID = ?", 1).First(&activeUser)
                if activeUser.UserID != 0 {
                  color.Yellow("You Are Already Logged In!")
                  return nil
                }
                developer := Developer{}
                username := cCtx.Args().First()
                password := cCtx.Args().Get(1)
                result:= db.Where("username = ?", username).First(&developer)
                if result.Error != nil{
                  color.Red("Invalid Username")
                  return nil
                }
                if developer.Password == password{
                  activeUser.UserID = developer.ID
                  activeUser.Role = "dev"
                  activeUser.Username =developer.Username
                  db.Save(&activeUser)
                  color.Green("Successfull Login.. Welcome "+developer.Username+"!")
                  return nil;
                }
                color.Red("Invalid Password")
                  return nil
              },
            },{
              Name:"delete",
              Usage: "<GameID>",
              Action: func(ctx *cli.Context) error {
                var activeUser ActiveUser
                db.Where("ID=?",1).First(&activeUser)
                if activeUser.Role == "user" {
                  color.Yellow("Cannot delete games as user.")
                  return nil
                }
                gameID := ctx.Args().First()
                var dev Developer
                db.Where("ID=?",activeUser.UserID).First(&dev)
                var game Game
                db.Where("ID=?",gameID).Find(&game)
                
                if game.DeveloperRefer != dev.ID{
                  color.Red("Cannot delete other's games!")
                  return nil
                }
                db.Delete(&game)
                color.Green("Game Deleted")
                return nil
              },
            },
            {
              Name: "signup",
              Usage: "<username> <password> <name>",
              Aliases: []string{"s"},
              Action: func(cCtx *cli.Context) error{
                var activeUser ActiveUser
                db.Where("ID = ?", 1).First(&activeUser)
                if activeUser.UserID != 0 {
                  color.Yellow("Can't signup while logged In!")
                  return nil
                }
                username:= cCtx.Args().First()
                password := cCtx.Args().Get(1)
                name := cCtx.Args().Get(2)
                if len(username) < 5{
                  color.Red("username length must be longer than 5 char")
                  return nil
                }
                if len(password) < 8 {
                  color.Red("password length must be longer than 8 char")
                  return nil
                }
                  result:= db.Create(&Developer{Username: cCtx.Args().First(),Password: cCtx.Args().Get(1),Name: name})
                  if int(result.RowsAffected) == 1 {
                  color.Green("Developer Created Successfully!")
                  return nil
                  }
                  color.Red("Something Went Wrong!")
                  return nil
              },
            },
          },
        },
        //LOGOUT
        {
          Name: "logout",
          Usage: "no parameter",
          Aliases: []string{"lo"},
          Action: func(cCtx *cli.Context) error{
            var activeUser ActiveUser
            db.Where("id = ?",1).First(&activeUser)
            if activeUser.UserID == 0 {
              color.Yellow("Not Logged In")
              return nil
            }

            activeUser.Role = "null"
            activeUser.UserID = 0
            activeUser.Username = "null"
            db.Save(activeUser)
            color.Green("Logged Out!")
            return nil
          },
        },
        //LIST
        {
          Name:"list",
          Usage: "lists",
          Subcommands: []*cli.Command{
            {
              Name:"genres",
              Action: func(ctx *cli.Context) error {
              var genres []Genre
                db.Find(&genres)
                
                fmt.Println("ID    NAME")
                for _,genre := range genres{
                  fmt.Println(genre.ID, "   ", genre.Name)
                }

                return nil
              },
            },{
              Name: "games",
              Usage: "List All Available Game",
              Action: func(ctx *cli.Context) error {
                var games []Game
                db.Find(&games)
                fmt.Println("ID Name                  PRICE        PUBLISHDATE")
                for _, game := range games{
                  fmt.Println(game.ID,"  ",game.Name,"   ",game.CurrentPrice,"        ",game.CreatedAt)
                }
                return nil
              },
            },{
              Name: "bygenre",
              Usage: "List games by genre",
              Action: func(ctx *cli.Context) error {
                genreID := ctx.Args().First()
                var games []Game
                db.Joins("JOIN game_genres on game_genres.game_id = games.id JOIN genres on game_genres.genre_id = genres.id AND genres.id = ? ", genreID).Group("games.id").Find(&games)
                fmt.Println(games)
                fmt.Println("ID    GAME                  PRICE        PUBLISHDATE")
                for _, game := range games{
                  fmt.Println(game.ID,"     ",game.Name,"   ",game.CurrentPrice,"        ",game.CreatedAt)
                }
                return nil
              },
            },
            {
              Name:"developers",
              Usage:"list all developers",
              Action: func(ctx *cli.Context) error {
                var developers []Developer
                db.Find(&developers)
                fmt.Println("ID   Name")
                for _,dev := range developers{
                  fmt.Println(dev.ID, "  ", dev.Name)
                }
                return nil 
              },
            },
            {
              Name:"bydev",
              Usage: "<developerID>",
              Action: func(ctx *cli.Context) error {
                var dev Developer
                db.Where("ID=?",ctx.Args().First()).Preload("Games").First(&dev)
                games := dev.Games

                fmt.Println("ID   NAME    PRICE")

                for _,game := range games{
                  fmt.Println(game.ID,"   ",game.Name,"    ",game.CurrentPrice)
                }


                return nil
              },
            },

          },
        },
        
    },
}
if err := app.Run(os.Args); err != nil {
  log.Fatal(err)
}

}