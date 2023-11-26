package game

import (
	"github.com/DTLP/mini_tanks/internal/actors"
	"github.com/DTLP/mini_tanks/internal/levels"

	"io"
	"io/ioutil"
	"fmt"
	"net"
	"time"
	"sync"
	"bytes"
	// "bufio"
	"strings"
	"encoding/json"
)

var (
	serverAddr  = "192.168.1.205:54050"
	globalConn  net.Conn
	Conn 	    net.Conn
	connMutex   sync.Mutex
	ClientInput string
	prevJsonTankState, prevJsonLevelNumUpdate, prevJsonLevelObjectUpdate []byte
)
var msg = ""

type levelNumMessage struct {
	Type     string `json:"Type"`
	LevelNum int    `json:"levelNum"`
}


func log(msg string) {
    fmt.Printf("[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), msg)
}

// Server
func hostNewCoopGame() {
	go startServer()
}

func startServer() {
    listener, err := net.Listen("tcp", serverAddr)
    if err != nil {
        // log("Error listening: " + err.Error())
        return
    }

    for {
        conn, err := listener.Accept()
        if err != nil {
            // log("Error accepting connection: " + err.Error())
            continue
        }

		// Lock the mutex before modifying the global connection variable
		connMutex.Lock()
		Conn = conn
		connMutex.Unlock()

        gameMode = 1
    }
}

func sendUpdatesToClient(tanks []actors.Tank) {
    go func() {
        // Send level num
        sendLevelNumtoClient(levelNum)

        // Create a copy of the slice
        tanksCopy := make([]actors.Tank, len(tanks))
        copy(tanksCopy, tanks)

        // Send tanks
        for i := 0; i < len(tanksCopy); i++ {
            if tanksCopy[i].Player != 2 {
                sendTankState(tanksCopy[i])
            }
        }



        levelObjectsCopy := make([]levels.LevelBlock, len(levelObjects))
        copy(levelObjectsCopy, levelObjects)
        
        for i := 0; i < len(levelObjectsCopy); i++ {
            sendLevelObjectToClient(levelObjectsCopy[i])
        }

        // Send level objects
        // for i := range levelObjects {
        //     sendLevelObjectToClient(levelObjects[i])
        // }
    }()
	
}

func receiveClientInput(Conn net.Conn) (string, error) {
    buffer := make([]byte, 1024)
    n, err := Conn.Read(buffer)
    if err != nil {
        if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
            // Handle timeout error, if needed
            // log("Timeout receiving client input")
            return "", nil
        }

        // Check for forcibly closed connection error
        if opErr, ok := err.(*net.OpError); ok {
            if opErr.Err.Error() == "\n ⚠️An existing connection was forcibly closed by the remote host." {
                // log("Client forcibly closed the connection")
                return "", io.EOF
            }
        }

        // log("Error receiving client input: " + err.Error())
        return "", err
    }

    if n == 0 {
        // log("Client disconnected")
        return "", io.EOF
    }

    input := string(buffer[:n])
    // log("Received client input: " + input)
    return input, nil
}

func handleClientInput() (actors.TankUpdate, error) {
	// Lock the mutex before accessing the global connection variable
	connMutex.Lock()
	defer connMutex.Unlock()

	clientInput, err := receiveClientInput(Conn)
	if err != nil {
		// Handle the error, e.g., client disconnected
		// log("Error receiving client input: " + err.Error())
		return actors.TankUpdate{}, err
	}

    if clientInput == "" {
        // Handle empty input or incomplete JSON
        return actors.TankUpdate{}, fmt.Errorf("\n ⚠️Empty or incomplete JSON input")
    }
	
    // fmt.Printf("\n Client input: %s\n", clientInput)


	var gameState actors.TankUpdate
    err = json.Unmarshal([]byte(clientInput), &gameState)
    if err != nil {
        // Handle decoding error
        fmt.Println("\n ⚠️ Error decoding input:", err)
        return actors.TankUpdate{}, err
    }

    return gameState, nil
}

func sendLevelNumtoClient(levelNum int) error {
	// Create a Message instance with the relevant parameters
	message := levelNumMessage{
		Type:     "levelNum",
		LevelNum: levelNum,
	}

    levelNumJSON, err := json.Marshal(message)
    if err != nil {
        // fmt.Printf("⚠️Error marshalling tank update:", err)
        return err
    }

	// fmt.Printf("\n ##### levelNum: %s",levelNumJSON)

    // Add a newline character between each JSON object
    levelNumJSONWithNewline := append(levelNumJSON, byte('\n'))

    // Compare current JSON update to previous JSON update
    if !bytes.Equal(levelNumJSONWithNewline, prevJsonLevelNumUpdate) {
        // fmt.Printf("\n➡️ Sending updated JSON: %s\n", tankUpdateJSONWithNewline)

        _, err = Conn.Write(levelNumJSONWithNewline)
        if err != nil {
            // fmt.Printf("⚠️Error sending tank update to server:", err)
            return err
        }

        // Update prevJsonGameState to the current update
        prevJsonLevelNumUpdate = levelNumJSONWithNewline
    } else {
        // fmt.Println("No changes in tank update. Skipping send.")
    }

    return nil
}

func sendLevelObjectToClient(object levels.LevelBlock) error {
    levelObjectUpdate := levels.ObjectUpdate{
        Type:         "levelObject",
        X:            object.X,
        Y:            object.Y,
        Width:        object.Width, 
        Height:       object.Height, 
        Blocks:       object.Blocks, 
        Health:       object.Health,
    }

    // Marshal the TankUpdate to JSON
    jsonLevelObjectUpdate, err := json.Marshal(levelObjectUpdate)
    if err != nil {
        fmt.Printf("⚠️ Error marshalling tank update:", err)
        return err
    }

    // Add a newline character between each JSON object
    jsonLevelObjectUpdateWithNewline := append(jsonLevelObjectUpdate, byte('\n'))

    // Compare current JSON update to previous JSON update
    if !bytes.Equal(jsonLevelObjectUpdateWithNewline, prevJsonLevelObjectUpdate) {
        // fmt.Printf("\n➡️ Sending updated JSON: %s\n", tankUpdateJSONWithNewline)

        _, err = Conn.Write(jsonLevelObjectUpdateWithNewline)
        if err != nil {
            // fmt.Printf("⚠️Error sending tank update to server:", err)
            return err
        }

        // Update prevJsonGameState to the current update
        prevJsonLevelObjectUpdate = jsonLevelObjectUpdateWithNewline
    } else {
        // fmt.Println("No changes in tank update. Skipping send.")
    }

    return nil
}

func processUpdatesFromClient(tanks *[]actors.Tank) {
    if updates, ok := getUpdates(); ok {
        for _, update := range updates {
            // Convert the update string to a map
            var updateMap map[string]interface{}
            if err := json.Unmarshal([]byte(update), &updateMap); err != nil {
                fmt.Println("\n ⚠️ Error decoding update:", err)
                continue
            }

            // Accessing specific fields based on their keys
            if updateType, ok := updateMap["Type"].(string); ok {
                // fmt.Println("\n Update Type:", updateType)
				// fmt.Println("\n Update:", update)

                // Perform actions based on the update type
                switch updateType {
				case "tank":

					updateTank(tanks, updateMap)
				// Add more cases as needed for different update types
				default:
					fmt.Println("\n ⚠️ Unknown update type:", updateType)
				}
			} else {
				fmt.Println("\n ⚠️ Update Type not found or not a string.")
			}

            
        }
    }
}


// Shared
func updateTank(tanks *[]actors.Tank, update map[string]interface{}) {
	name := update["Name"].(string)

	// Find the tank in the list based on the name
	for i, t := range *tanks {
		if t.Name == name {
			// fmt.Println("💚 t.Name %s:", t.Name)

			// Update the tank properties
			(*tanks)[i].X = update["X"].(float64)
			(*tanks)[i].Y = update["Y"].(float64)
			(*tanks)[i].Hull.Angle = update["HullAngle"].(float64)
			(*tanks)[i].Hull.CollisionX1 = update["X1"].(float64)
			(*tanks)[i].Hull.CollisionY1 = update["Y1"].(float64)
			(*tanks)[i].Hull.CollisionX2 = update["X2"].(float64)
			(*tanks)[i].Hull.CollisionY2 = update["Y2"].(float64)
			(*tanks)[i].Hull.CollisionX3 = update["X3"].(float64)
			(*tanks)[i].Hull.CollisionY3 = update["Y3"].(float64)
			(*tanks)[i].Hull.CollisionX4 = update["X4"].(float64)
			(*tanks)[i].Hull.CollisionY4 = update["Y4"].(float64)
			(*tanks)[i].Turret.Angle = update["TurretAngle"].(float64)

			// Type conversion for Projectiles
			projectiles, err := convertToProjectileSlice(update["Projectiles"])
			if err != nil {
				fmt.Println("\n ⚠️ Error:", err)
				return
			}
			(*tanks)[i].Projectiles = projectiles

			// Type conversion for Explosions
			explosions, err := convertToExplosionSlice(update["Explosions"])
			if err != nil {
				fmt.Println("\n ⚠️ Error:", err)
				return
			}
			(*tanks)[i].Explosions = explosions

			// Break the loop since we found and updated the tank
			break
		}
	}
}

func convertToProjectileSlice(slice interface{}) ([]actors.Projectile, error) {
    if slice == nil {
        return nil, nil
    }

    projectilesData, ok := slice.([]interface{})
    if !ok {
        return nil, fmt.Errorf("\n ⚠️unable to convert to []interface{}")
    }

    result := make([]actors.Projectile, len(projectilesData))
    for i, p := range projectilesData {
        projectileBytes, err := json.Marshal(p)
        if err != nil {
            return nil, fmt.Errorf("\n ⚠️unable to convert Projectile to the expected type")
        }

        var projectile actors.Projectile
        err = json.Unmarshal(projectileBytes, &projectile)
        if err != nil {
            return nil, fmt.Errorf("\n ⚠️unable to convert Projectile to the expected type")
        }

        result[i] = projectile
    }

    return result, nil
}

func convertToExplosionSlice(slice interface{}) ([]actors.Explosion, error) {
    if slice == nil {
        return nil, nil
    }

    explosionsData, ok := slice.([]interface{})
    if !ok {
        return nil, fmt.Errorf("\n ⚠️ unable to convert to []interface{}")
    }

    result := make([]actors.Explosion, len(explosionsData))
    for i, e := range explosionsData {
        explosionBytes, err := json.Marshal(e)
        if err != nil {
            return nil, fmt.Errorf("\n ⚠️unable to convert Explosion to the expected type")
        }

        var explosion actors.Explosion
        err = json.Unmarshal(explosionBytes, &explosion)
        if err != nil {
            return nil, fmt.Errorf("\n ⚠️unable to convert Explosion to the expected type")
        }

        result[i] = explosion
    }

    return result, nil
}

func sendTankState(tank actors.Tank) error {
    // Create a TankUpdate instance with the relevant parameters
    tankUpdate := actors.TankUpdate{
		Type:            "tank",
        X:        	     tank.X,
		Y:		  	     tank.Y,
		Name:            tank.Name,
		HullAngle:		 tank.Hull.Angle,
		X1: tank.Hull.CollisionX1,
		Y1: tank.Hull.CollisionY1,
		X2: tank.Hull.CollisionX2,
		Y2: tank.Hull.CollisionY2,
		X3: tank.Hull.CollisionX3,
		Y3: tank.Hull.CollisionY3,
		X4: tank.Hull.CollisionX4,
		Y4: tank.Hull.CollisionY4,
        TurretAngle:     tank.Turret.Angle,
        Projectiles:     tank.Projectiles,
		Explosions:      tank.Explosions,
    }

    // Marshal the TankUpdate to JSON
    tankUpdateJSON, err := json.Marshal(tankUpdate)
    if err != nil {
        // fmt.Printf("⚠️Error marshalling tank update:", err)
        return err
    }

    // Add a newline character between each JSON object
    tankUpdateJSONWithNewline := append(tankUpdateJSON, byte('\n'))

    // Compare current JSON update to previous JSON update
    if !bytes.Equal(tankUpdateJSONWithNewline, prevJsonTankState) {
        // If there are changes, send only the changes
        // fmt.Printf("\n➡️ Sending updated JSON: %s\n", tankUpdateJSONWithNewline)

        _, err = Conn.Write(tankUpdateJSONWithNewline)
        if err != nil {
            // fmt.Printf("⚠️Error sending tank update to server:", err)
            return err
        }

        // Update prevJsonTankState to the current update
        prevJsonTankState = tankUpdateJSONWithNewline
    } else {
        // fmt.Println("No changes in tank update. Skipping send.")
    }

    return nil
}

const bufferSize = 4096 // Adjust the buffer size as needed
var jsonBuffer string
func receiveClientInputWithBuffer(conn net.Conn) error {
    buffer := make([]byte, bufferSize)
    bytesRead, err := conn.Read(buffer)
    if err != nil {
        return err
    }
    jsonBuffer += string(buffer[:bytesRead])
    return nil
}

func getUpdates() ([]string, bool) {
    connMutex.Lock()
    defer connMutex.Unlock()

    err := receiveClientInputWithBuffer(Conn)
    if err != nil {
        fmt.Println("\n ⚠️ Error receiving client input:", err)
        return nil, false
    }

	// fmt.Printf("\n----\nBUFFER: %s", jsonBuffer)

    // Split the string into individual JSON objects
    var updates []string
    decoder := json.NewDecoder(strings.NewReader(jsonBuffer))
    for {
        var update map[string]interface{}
        if err := decoder.Decode(&update); err == io.EOF {
            // Handle EOF (end of stream) appropriately
            break
        } else if err != nil {
            if err == io.ErrUnexpectedEOF {
                // Incomplete JSON, break and wait for more data
                break
            } else {
                fmt.Println("\n ⚠️ Error decoding JSON update:", err)
                fmt.Printf("\n🍗 Raw update: %s", jsonBuffer)
                return nil, false
            }
        }

        // Convert the map back to JSON and append to the updates slice
        updateJSON, err := json.Marshal(update)
        if err != nil {
            fmt.Println("\n ⚠️ Error encoding JSON update:", err)
            return nil, false
        }
        updates = append(updates, string(updateJSON))
    }

    // Update the jsonBuffer with any remaining unprocessed JSON
    remainingJSON, err := ioutil.ReadAll(decoder.Buffered())
    if err != nil {
        fmt.Println("\n ⚠️ Error reading remaining JSON:", err)
        return nil, false
    }
    jsonBuffer = string(remainingJSON)

	// fmt.Printf("\n----\nBUFFER: %s", jsonBuffer)
	// fmt.Printf("\n=============\n")

	// fmt.Println("\n🦬 jsonBuffer length:", len(jsonBuffer))

    return updates, true

//////////// Todo: investigate issues with lag and desync //////////////////////
}


// Client
func joinNewCoopGame() {
	go joinServer()
}

func joinServer() {
    conn, err := net.Dial("tcp", serverAddr)
    if err != nil {
        // log("Error connecting: " + err.Error())
        return
    }
	Conn = conn

    becomePlayer2()
}

func updateLevelObjects(levelObjects *[]levels.LevelBlock, update map[string]interface{}) {
    for i, object := range *levelObjects {
        if object.X == update["X"].(float64) && object.Y == update["Y"].(float64) {
            // Update the level object properties
            (*levelObjects)[i].Width = update["Width"].(float64)
            (*levelObjects)[i].Height = update["Height"].(float64)
            (*levelObjects)[i].Blocks = buildBlocks(update["Blocks"].([]interface{}))
            (*levelObjects)[i].Health = int(update["Health"].(float64))
        }
    }
}

func buildBlocks(blocksData []interface{}) []levels.Block {
    blocks := make([]levels.Block, len(blocksData))
    for i, blockData := range blocksData {
        blockMap, ok := blockData.(map[string]interface{})
        if !ok {
            // Handle the case where the type conversion fails
            fmt.Println("\n ⚠️ Error: Unable to convert blockData to map[string]interface{}")
            continue
        }

        wallsData, ok := blockMap["Walls"].([]interface{})
        if !ok {
            // Handle the case where the type conversion fails
            fmt.Println("\n ⚠️ Error: Unable to convert Walls to []interface{}")
            continue
        }

        walls := make([]levels.Line, len(wallsData))
        for j, wallData := range wallsData {
            wallMap, ok := wallData.(map[string]interface{})
            if !ok {
                // Handle the case where the type conversion fails
                fmt.Println("\n ⚠️ Error: Unable to convert wallData to map[string]interface{}")
                continue
            }

            line := levels.Line{
                X1: wallMap["X1"].(float64),
                Y1: wallMap["Y1"].(float64),
                X2: wallMap["X2"].(float64),
                Y2: wallMap["Y2"].(float64),
            }

            walls[j] = line
        }

        block := levels.Block{
            Walls: walls,
        }

        blocks[i] = block
    }
    return blocks
}


func processUpdatesFromServer(tanks *[]actors.Tank, levelNum *int, levelObjects *[]levels.LevelBlock) {
    // ...

    if updates, ok := getUpdates(); ok {
        for _, update := range updates {
            // Convert the update string to a map
            var updateMap map[string]interface{}
            if err := json.Unmarshal([]byte(update), &updateMap); err != nil {
                fmt.Println("\n ⚠️ Error decoding update:", err)
				fmt.Println("\n ⚠️ update: %s\n", update)
                continue
            }


            // Accessing specific fields based on their keys
            if updateType, ok := updateMap["Type"].(string); ok {
                // fmt.Println("\n Update Type:", updateType)
				// fmt.Println("\n Update:", update)

                // Perform actions based on the update type
                switch updateType {
				case "tank":
					updateTank(tanks, updateMap)

				case "levelNum":
					if *levelNum != int(updateMap["levelNum"].(float64)) {
						*levelNum = int(updateMap["levelNum"].(float64))
						*levelObjects = levels.GetLevelObjects(int(updateMap["levelNum"].(float64)))
                        actors.ResetPlayerPositions(tanks)
					}

				case "levelObject":
					updateLevelObjects(levelObjects, updateMap)

				default:
					fmt.Println("\n ⚠️ Unknown update type:", updateType)
				}
			} else {
				fmt.Println("\n ⚠️ Update Type not found or not a string.")
			}

            
        }
    }
}


// Coop logic
func becomePlayer2() {
	player   = 2
	gameMode = 1
}


func doesPlayer2Exist(tanks []actors.Tank) bool {
    for _, tank := range tanks {
        if tank.Name == "player2" {
            return true
        }
    }

    return false
}