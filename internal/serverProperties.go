package internal

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
)

type ServerProperties struct {
	// Used as the server name
	// Allowed values: Any string without semicolon symbol.
	ServerName string

	// Sets the game mode for new players.
	// Allowed values: "survival", "creative", or "adventure"
	GameMode string

	// force-gamemode=false (or force-gamemode is not defined in the server.properties)
	// prevents the server from sending to the client gamemode values other
	// than the gamemode value saved by the server during world creation
	// even if those values are set in server.properties after world creation.
	//
	// force-gamemode=true forces the server to send to the client gamemode values
	// other than the gamemode value saved by the server during world creation
	// if those values are set in server.properties after world creation.
	ForceGameMode bool

	// Sets the difficulty of the world.
	// Allowed values: "peaceful", "easy", "normal", or "hard"
	Difficulty string

	// If true then cheats like commands can be used.
	// Allowed values: "true" or "false"
	AllowCheats bool

	// The maximum number of players that can play on the server.
	// Allowed values: Any positive integer
	MaxPlayers int

	// If true then all connected players must be authenticated to Xbox Live.
	// Clients connecting to remote (non-LAN) servers will always require Xbox Live authentication regardless of this setting.
	// If the server accepts connections from the Internet, then it's highly recommended to enable online-mode.
	// Allowed values: "true" or "false"
	OnlineMode bool

	// If true then all connected players must be listed in the separate allowlist.json file.
	// Allowed values: "true" or "false"
	AllowList bool

	// Which IPv4 port the server should listen to.
	// Allowed values: Integers in the range [1, 65535]
	ServerPort int

	// Which IPv6 port the server should listen to.
	// Allowed values: Integers in the range [1, 65535]
	ServerPortV6 int

	// Listen and respond to clients that are looking for servers on the LAN. This will cause the server
	// to bind to the default ports (19132, 19133) even when `server-port` and `server-portv6`
	// have non-default values. Consider turning this off if LAN discovery is not desirable, or when
	// running multiple servers on the same host may lead to port conflicts.
	// Allowed values: "true" or "false"
	EnableLanVisibility bool

	// The maximum allowed view distance in number of chunks.
	// Allowed values: Positive integer equal to 5 or greater.
	ViewDistance int

	// The world will be ticked this many chunks away from any player.
	// Allowed values: Integers in the range [4, 12]
	TickDistance int

	// After a player has idled for this many minutes they will be kicked. If set to 0 then players can idle indefinitely.
	// Allowed values: Any non-negative integer.
	PlayerIdleTimeout int

	// Maximum number of threads the server will try to use. If set to 0 or removed then it will use as many as possible.
	// Allowed values: Any positive integer.
	MaxThreads int

	// Allowed values: Any string without semicolon symbol or symbols illegal for file name: /\n\r\t\f`?*\\<>|\":
	LevelName string

	// Use to randomize the world
	// Allowed values: Any string
	LevelSeed string

	// Permission level for new players joining for the first time.
	// Allowed values: "visitor", "member", "operator"
	DefaultPlayerPermissionLevel string

	// Force clients to use texture packs in the current world
	// Allowed values: "true" or "false"
	TexturePackRequired bool

	// Enables logging content errors to a file
	// Allowed values: "true" or "false"
	ContentLogFileEnabled bool

	// Determines the smallest size of raw network payload to compress
	// Allowed values: 0-65535
	CompressionThreshold int

	// Determines the compression algorithm to use for networking
	// Allowed values: "zlib", "snappy"
	CompressionAlgorithm string

	// Allowed values: "client-auth", "server-auth", "server-auth-with-rewind"
	// Enables server authoritative movement. If "server-auth", the server will replay local user input on
	// the server and send down corrections when the client's position doesn't match the server's.
	// If "server-auth-with-rewind" is enabled and the server sends a correction, the clients will be instructed
	// to rewind time back to the correction time, apply the correction, then replay all the player's inputs since then. This results in smoother and more frequent corrections.
	// Corrections will only happen if correct-player-movement is set to true.
	ServerAuthoritativeMovement string

	// The number of incongruent time intervals needed before abnormal behavior is reported.
	// Disabled by server-authoritative-movement.
	PlayerMovementScoreThreshold int

	// The amount that the player's attack direction and look direction can differ.
	// Allowed values: Any value in the range of [0, 1] where 1 means that the
	// direction of the players view and the direction the player is attacking
	// must match exactly and a value of 0 means that the two directions can
	// differ by up to and including 90 degrees.
	PlayerMovementActionDirectionThreshold float64

	// The difference between server and client positions that needs to be exceeded before abnormal behavior is detected.
	// Disabled by server-authoritative-movement.
	PlayerMovementDistanceThreshold float64

	// The duration of time the server and client positions can be out of sync (as defined by player-movement-distance-threshold)
	// before the abnormal movement score is incremented. This value is defined in milliseconds.
	// Disabled by server-authoritative-movement.
	PlayerMovementDurationThresholdInMs int

	// If true, the client position will get corrected to the server position if the movement score exceeds the threshold.
	CorrectPlayerMovement bool

	// If true, the server will compute block mining operations in sync with the client so it can verify that the client should be able to break blocks when it thinks it can.
	ServerAuthoritativeBlockBreaking bool

	// Allowed values: "None", "Dropped", "Disabled"
	// This represents the level of restriction applied to the chat for each player that joins the server.
	// "None" is the default and represents regular free chat.
	// "Dropped" means the chat messages are dropped and never sent to any client. Players receive a message to let them know the feature is disabled.
	// "Disabled" means that unless the player is an operator, the chat UI does not even appear. No information is displayed to the player.
	ChatRestriction string

	// If true, the server will inform clients that they should ignore other players when interacting with the world. This is not server authoritative.
	DisablePlayerInteraction bool

	// If true, the server will inform clients that they have the ability to generate visual level chunks outside of player interaction distances.
	ClientSideChunkGenerationEnabled bool

	// If true, the server will send hashed block network ID's instead of id's that start from 0 and go up.  These id's are stable and won't change regardless of other block changes.
	BlockNetworkIdsAreHashes bool

	// Internal Use Only
	DisablePersona bool

	// If true, disable players customized skins that were customized outside of the Minecraft store assets or in game assets.  This is used to disable possibly offensive custom skins players make.
	DisableCustomSkins bool

	// Allowed values: "Disabled" or any value in range [0.0, 1.0]
	// If "Disabled" the server will dynamically calculate how much of the player's view it will generate, assigning the rest to the client to build.
	// Otherwise from the overridden ratio tell the server how much of the player's view to generate, disregarding client hardware capability.
	// Only valid if client-side-chunk-generation-enabled is enabled
	ServerBuildRadiusRatio string

	AppName string

	AppVersion string
}

func (s *ServerProperties) Template() string {
	s.AppName = "AppName"
	s.AppVersion = "AppVersion"

	tmplFile := "templates/1.20.20.tmpl"
	_, err := os.Stat(tmplFile)
	if err != nil {
		fmt.Println("Error: ", err)
		panic(err)
	}

	template, err := template.ParseFiles(tmplFile)
	// Capture any error
	if err != nil {
		log.Fatalln(err)
	}

	var buffer bytes.Buffer
	// Print out the template to std
	template.Execute(&buffer, s)
	return buffer.String()
}

func NewServerProperties() *ServerProperties {
	return &ServerProperties{
		ServerName:                             "Minecraft Server",
		GameMode:                               "survival",
		ForceGameMode:                          false,
		Difficulty:                             "easy",
		AllowCheats:                            false,
		MaxPlayers:                             10,
		OnlineMode:                             true,
		AllowList:                              false,
		ServerPort:                             19132,
		ServerPortV6:                           19133,
		EnableLanVisibility:                    true,
		ViewDistance:                           10,
		TickDistance:                           4,
		PlayerIdleTimeout:                      30,
		MaxThreads:                             8,
		LevelName:                              "Bedrock level",
		LevelSeed:                              "",
		DefaultPlayerPermissionLevel:           "visitor",
		TexturePackRequired:                    false,
		ContentLogFileEnabled:                  false,
		CompressionThreshold:                   1,
		CompressionAlgorithm:                   "zlib",
		ServerAuthoritativeMovement:            "server-auth",
		PlayerMovementScoreThreshold:           20,
		PlayerMovementActionDirectionThreshold: 0.0625,
		PlayerMovementDistanceThreshold:        0.03125,
		PlayerMovementDurationThresholdInMs:    500,
		CorrectPlayerMovement:                  true,
		ServerAuthoritativeBlockBreaking:       true,
		ChatRestriction:                        "None",
		DisablePlayerInteraction:               false,
		ClientSideChunkGenerationEnabled:       true,
		BlockNetworkIdsAreHashes:               false,
		DisablePersona:                         false,
		DisableCustomSkins:                     false,
		ServerBuildRadiusRatio:                 "Disabled",
	}
}
