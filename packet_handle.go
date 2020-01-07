package typhoon

import (
	"encoding/binary"
	"fmt"
	"github.com/TyphoonMC/go.uuid"
	"log"
)

type PacketHandshake struct {
	Protocol Protocol
	Address  string
	Port     uint16
	State    State
}

func (packet *PacketHandshake) Read(player *Player, length int) (err error) {
	protocol, err := player.ReadVarInt()
	if err != nil {
		log.Print(err)
		return
	}
	packet.Protocol = Protocol(protocol)
	packet.Address, err = player.ReadStringLimited(config.BufferConfig.HandshakeAddress)
	if err != nil {
		log.Print(err)
		return
	}
	packet.Port, err = player.ReadUInt16()
	if err != nil {
		log.Print(err)
		return
	}
	state, err := player.ReadVarInt()
	if err != nil {
		log.Print(err)
		return
	}
	packet.State = State(state)
	return
}
func (packet *PacketHandshake) Write(player *Player) (err error) {
	return
}
func (packet *PacketHandshake) Handle(player *Player) {
	player.state = packet.State
	player.protocol = packet.Protocol
	player.inaddr.address = packet.Address
	player.inaddr.port = packet.Port
}
func (packet *PacketHandshake) Id() (int, Protocol) {
	return 0x00, V1_10
}

type PacketStatusRequest struct{}

func (packet *PacketStatusRequest) Read(player *Player, length int) (err error) {
	return
}
func (packet *PacketStatusRequest) Write(player *Player) (err error) {
	return
}
func (packet *PacketStatusRequest) Handle(player *Player) {
	protocol := COMPATIBLE_PROTO[0]
	if IsCompatible(player.protocol) {
		protocol = player.protocol
	}

	max_players := config.MaxPlayers
	motd := config.Motd

	count := player.core.playerRegistry.GetPlayerCount()
	if max_players < count && !config.Restricted {
		max_players = count
	}

	response := PacketStatusResponse{
		Response: fmt.Sprintf(`{"version":{"name":"Typhoon","protocol":%d},"players":{"max":%d,"online":%d,"sample":[]},"description":{"text":"%s"},"favicon":"%s","modinfo":{"type":"FML","modList":[]}}`, protocol, max_players, count, JsonEscape(motd), JsonEscape(favicon)),
	}
	player.WritePacket(&response)
}
func (packet *PacketStatusRequest) Id() (int, Protocol) {
	return 0x00, V1_10
}

type PacketStatusResponse struct {
	Response string
}

func (packet *PacketStatusResponse) Read(player *Player, length int) (err error) {
	return
}
func (packet *PacketStatusResponse) Write(player *Player) (err error) {
	err = player.WriteString(packet.Response)
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketStatusResponse) Handle(player *Player) {}
func (packet *PacketStatusResponse) Id() (int, Protocol) {
	return 0x00, V1_10
}

type PacketStatusPing struct {
	Time uint64
}

func (packet *PacketStatusPing) Read(player *Player, length int) (err error) {
	packet.Time, err = player.ReadUInt64()
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketStatusPing) Write(player *Player) (err error) {
	err = player.WriteUInt64(packet.Time)
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketStatusPing) Handle(player *Player) {
	player.WritePacket(packet)
}
func (packet *PacketStatusPing) Id() (int, Protocol) {
	return 0x01, V1_10
}

type PacketLoginStart struct {
	Username string
}

func (packet *PacketLoginStart) Read(player *Player, length int) (err error) {
	packet.Username, err = player.ReadStringLimited(config.BufferConfig.PlayerName)
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketLoginStart) Write(player *Player) (err error) {
	return
}

var (
	join_game = PacketPlayJoinGame{
		EntityId:            0,
		Gamemode:            SPECTATOR,
		Dimension:           END,
		HashedSeed:          0,
		Difficulty:          NORMAL,
		LevelType:           DEFAULT,
		MaxPlayers:          0xFF,
		ReducedDebug:        false,
		EnableRespawnScreen: true,
	}
	position_look = PacketPlayerPositionLook{}
)

func (packet *PacketLoginStart) Handle(player *Player) {
	if !IsCompatible(player.protocol) {
		player.Kick("Incompatible version")
		return
	}

	max_players := config.MaxPlayers

	count := player.core.playerRegistry.GetPlayerCount()
	if max_players <= count && config.Restricted {
		player.Kick("Server is full")
	}

	player.name = packet.Username

	if config.Compression && player.protocol >= V1_8 {
		setCompression := PacketSetCompression{config.Threshold}
		player.WritePacket(&setCompression)
		player.compression = true
	}

	success := PacketLoginSuccess{
		UUID:     player.uuid,
		Username: player.name,
	}
	player.WritePacket(&success)
	player.state = PLAY
	player.register()

	player.WritePacket(&join_game)
	player.WritePacket(&position_look)

	if player.protocol >= V1_13 {
		player.WritePacket(&PacketPlayDeclareCommands{
			player.core.compiledCommands,
			0,
		})
	}

	player.core.CallEvent(&PlayerJoinEvent{
		player,
	})
}
func (packet *PacketLoginStart) Id() (int, Protocol) {
	return 0x00, V1_10
}

type PacketLoginDisconnect struct {
	Component string
}

func (packet *PacketLoginDisconnect) Read(player *Player, length int) (err error) {
	return
}
func (packet *PacketLoginDisconnect) Write(player *Player) (err error) {
	err = player.WriteString(packet.Component)
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketLoginDisconnect) Handle(player *Player) {}
func (packet *PacketLoginDisconnect) Id() (int, Protocol) {
	return 0x00, V1_10
}

type PacketLoginSuccess struct {
	UUID     string
	Username string
}

func (packet *PacketLoginSuccess) Read(player *Player, length int) (err error) {
	return
}
func (packet *PacketLoginSuccess) Write(player *Player) (err error) {
	err = player.WriteString(packet.UUID)
	if err != nil {
		log.Print(err)
		return
	}
	err = player.WriteString(packet.Username)
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketLoginSuccess) Handle(player *Player) {}
func (packet *PacketLoginSuccess) Id() (int, Protocol) {
	return 0x02, V1_10
}

type PacketSetCompression struct {
	Threshold int
}

func (packet *PacketSetCompression) Read(player *Player, length int) (err error) {
	return
}
func (packet *PacketSetCompression) Write(player *Player) (err error) {
	err = player.WriteVarInt(packet.Threshold)
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketSetCompression) Handle(player *Player) {}
func (packet *PacketSetCompression) Id() (int, Protocol) {
	return 0x03, V1_10
}

type PacketPlayChat struct {
	Message string
}

func (packet *PacketPlayChat) Read(player *Player, length int) (err error) {
	packet.Message, err = player.ReadStringLimited(config.BufferConfig.ChatMessage)
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketPlayChat) Write(player *Player) (err error) {
	return
}
func (packet *PacketPlayChat) Handle(player *Player) {
	if len(packet.Message) > 0 {
		if packet.Message[0] != '/' {
			player.core.CallEvent(&PlayerChatEvent{
				player,
				packet.Message,
			})
		} else {
			player.core.onCommand(player, packet.Message[1:len(packet.Message)])
		}
	}
}
func (packet *PacketPlayChat) Id() (int, Protocol) {
	return 0x02, V1_10
}

type PacketPlayTabComplete struct {
	Matches []string
}

func (packet *PacketPlayTabComplete) Read(player *Player, length int) (err error) {
	return
}
func (packet *PacketPlayTabComplete) Write(player *Player) (err error) {
	err = player.WriteVarInt(len(packet.Matches))
	if err != nil {
		log.Print(err)
		return
	}
	for _, s := range packet.Matches {
		err = player.WriteString(s)
		if err != nil {
			log.Print(err)
			return
		}
	}
	return
}
func (packet *PacketPlayTabComplete) Handle(player *Player) {}
func (packet *PacketPlayTabComplete) Id() (int, Protocol) {
	return 0x0E, V1_10
}

type PacketPlayTabCompleteServerbound struct {
	Text          string
	AssumeCommand bool
	Position      Position
}

func (packet *PacketPlayTabCompleteServerbound) Read(player *Player, length int) (err error) {
	packet.Text, err = player.ReadStringLimited(config.BufferConfig.ChatMessage)
	if err != nil {
		log.Print(err)
		return
	}
	packet.AssumeCommand, err = player.ReadBool()
	if err != nil {
		log.Print(err)
		return
	}
	hasPosition, err := player.ReadBool()
	if err != nil {
		log.Print(err)
		return
	}
	if hasPosition {
		packet.Position, err = player.ReadPosition()
		if err != nil {
			log.Print(err)
			return
		}
	}
	return
}
func (packet *PacketPlayTabCompleteServerbound) Write(player *Player) (err error) {
	return
}
func (packet *PacketPlayTabCompleteServerbound) Handle(player *Player) {
	if len(packet.Text) > 0 {
		if packet.Text[0] == '/' {
			player.core.onTabCommand(player, packet.Text[1:len(packet.Text)])
		}
	}
}
func (packet *PacketPlayTabCompleteServerbound) Id() (int, Protocol) {
	return 0x01, V1_10
}

type PacketPlayClientStatus struct {
	Action ClientStatusAction
}

func (packet *PacketPlayClientStatus) Read(player *Player, length int) (err error) {
	act, err := player.ReadVarInt()
	if err != nil {
		log.Print(err)
		return
	}
	packet.Action = ClientStatusAction(act)
	return
}
func (packet *PacketPlayClientStatus) Write(player *Player) (err error) {
	return
}
func (packet *PacketPlayClientStatus) Handle(player *Player) {
	return
}
func (packet *PacketPlayClientStatus) Id() (int, Protocol) {
	return 0x03, V1_10
}

type PacketPlayMessage struct {
	Component string
	Position  ChatPosition
}

func (packet *PacketPlayMessage) Read(player *Player, length int) (err error) {
	return
}
func (packet *PacketPlayMessage) Write(player *Player) (err error) {
	err = player.WriteString(packet.Component)
	if err != nil {
		log.Print(err)
		return
	}
	if player.protocol > V1_7_6 {
		err = player.WriteUInt8(uint8(packet.Position))
		if err != nil {
			log.Print(err)
			return
		}
	}
	return
}
func (packet *PacketPlayMessage) Handle(player *Player) {}
func (packet *PacketPlayMessage) Id() (int, Protocol) {
	return 0x0F, V1_10
}

type PacketBossBar struct {
	UUID     uuid.UUID
	Action   BossBarAction
	Title    string
	Health   float32
	Color    BossBarColor
	Division BossBarDivision
	Flags    uint8
}

func (packet *PacketBossBar) Read(player *Player, length int) (err error) {
	return
}
func (packet *PacketBossBar) Write(player *Player) (err error) {
	err = player.WriteUUID(packet.UUID)
	if err != nil {
		log.Print(err)
		return
	}
	err = player.WriteVarInt(int(packet.Action))
	if err != nil {
		log.Print(err)
		return
	}
	if packet.Action == BOSSBAR_UPDATE_TITLE || packet.Action == BOSSBAR_ADD {
		err = player.WriteString(packet.Title)
		if err != nil {
			log.Print(err)
			return
		}
	}
	if packet.Action == BOSSBAR_UPDATE_HEALTH || packet.Action == BOSSBAR_ADD {
		err = player.WriteFloat32(packet.Health)
		if err != nil {
			log.Print(err)
			return
		}
	}
	if packet.Action == BOSSBAR_UPDATE_STYLE || packet.Action == BOSSBAR_ADD {
		err = player.WriteVarInt(int(packet.Color))
		if err != nil {
			log.Print(err)
			return
		}
		err = player.WriteVarInt(int(packet.Division))
		if err != nil {
			log.Print(err)
			return
		}
	}
	if packet.Action == BOSSBAR_UPDATE_STYLE || packet.Action == BOSSBAR_ADD {
		err = player.WriteUInt8(packet.Flags)
		if err != nil {
			log.Print(err)
			return
		}
	}
	return
}
func (packet *PacketBossBar) Handle(player *Player) {}
func (packet *PacketBossBar) Id() (int, Protocol) {
	return 0x0C, V1_10
}

type PacketPlayDeclareCommands struct {
	Nodes     []commandNode
	RootIndex int
}

func (packet *PacketPlayDeclareCommands) Read(player *Player, length int) (err error) {
	return
}
func (packet *PacketPlayDeclareCommands) Write(player *Player) (err error) {
	err = player.WriteVarInt(len(packet.Nodes))
	if err != nil {
		log.Print(err)
		return
	}
	for _, n := range packet.Nodes {
		err = (&n).writeTo(player)
		if err != nil {
			log.Print(err)
			return
		}
	}
	err = player.WriteVarInt(packet.RootIndex)
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketPlayDeclareCommands) Handle(player *Player) {}
func (packet *PacketPlayDeclareCommands) Id() (int, Protocol) {
	return 0x11, V1_13
}

type PacketPlayPluginMessage struct {
	Channel string
	Data    []byte
}

func (packet *PacketPlayPluginMessage) Read(player *Player, length int) (err error) {
	var read int
	packet.Channel, read, err = player.ReadNStringLimited(20)
	if err != nil {
		log.Print(err)
		return
	}

	dataLength := length - read
	if player.protocol < V1_8 {
		sread, err := player.ReadUInt16()
		if err != nil {
			log.Print(err)
			return err
		}
		dataLength = int(sread)
	}

	packet.Data, err = player.ReadByteArray(dataLength)
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketPlayPluginMessage) Write(player *Player) (err error) {
	err = player.WriteString(packet.Channel)
	if err != nil {
		log.Print(err)
		return
	}
	if player.protocol < V1_8 {
		err = player.WriteUInt16(uint16(len(packet.Data)))
		if err != nil {
			log.Print(err)
			return err
		}
	}
	err = player.WriteByteArray(packet.Data)
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketPlayPluginMessage) Handle(player *Player) {
	if packet.Channel == "MC|Brand" || packet.Channel == "minecraft:brand" {
		log.Printf("%s is using %s client", player.name, string(packet.Data))
		buff := make([]byte, len(player.core.brand)+1)
		length := binary.PutUvarint(buff, uint64(len(player.core.brand)))
		copy(buff[length:], []byte(player.core.brand))
		player.WritePacket(&PacketPlayPluginMessage{
			packet.Channel,
			buff,
		})
	}
	player.core.CallEvent(&PluginMessageEvent{
		packet.Channel,
		packet.Data,
	})
}
func (packet *PacketPlayPluginMessage) Id() (int, Protocol) {
	return 0x18, V1_10
}

type PacketPlayDisconnect struct {
	Component string
}

func (packet *PacketPlayDisconnect) Read(player *Player, length int) (err error) {
	return
}
func (packet *PacketPlayDisconnect) Write(player *Player) (err error) {
	err = player.WriteString(packet.Component)
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketPlayDisconnect) Handle(player *Player) {}
func (packet *PacketPlayDisconnect) Id() (int, Protocol) {
	return 0x1A, V1_10
}

type PacketPlayKeepAlive struct {
	Identifier int
}

func (packet *PacketPlayKeepAlive) Read(player *Player, length int) (err error) {
	if player.protocol >= V1_12_2 {
		id, stt := player.ReadUInt64()
		packet.Identifier = int(id)
		err = stt
	} else if player.protocol <= V1_7_6 {
		id, stt := player.ReadUInt32()
		packet.Identifier = int(id)
		err = stt
	} else {
		packet.Identifier, err = player.ReadVarInt()
	}
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketPlayKeepAlive) Write(player *Player) (err error) {
	if player.protocol >= V1_12_2 {
		err = player.WriteUInt64(uint64(packet.Identifier))
	} else if player.protocol <= V1_7_6 {
		err = player.WriteUInt32(uint32(packet.Identifier))
	} else {
		err = player.WriteVarInt(packet.Identifier)
	}
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketPlayKeepAlive) Handle(player *Player) {
	if player.protocol > V1_8 {
		if player.keepalive != packet.Identifier {
			player.Kick("Invalid keepalive")
		}
	} else {
		player.keepalive = packet.Identifier
	}
	player.keepalive = 0
}
func (packet *PacketPlayKeepAlive) Id() (int, Protocol) {
	return 0x1F, V1_10
}

type PacketPlayJoinGame struct {
	EntityId            uint32
	Gamemode            Gamemode
	Dimension           Dimension
	HashedSeed          uint64
	Difficulty          Difficulty
	MaxPlayers          uint8
	LevelType           LevelType
	ReducedDebug        bool
	EnableRespawnScreen bool
}

func (packet *PacketPlayJoinGame) Read(player *Player, length int) (err error) {
	return
}
func (packet *PacketPlayJoinGame) Write(player *Player) (err error) {
	if player.protocol <= V1_9 {
		err = player.WriteUInt8(uint8(packet.EntityId))
	} else {
		err = player.WriteUInt32(packet.EntityId)
	}
	if err != nil {
		log.Print(err)
		return
	}
	err = player.WriteUInt8(uint8(packet.Gamemode))
	if err != nil {
		log.Print(err)
		return
	}
	err = player.WriteUInt32(uint32(packet.Dimension))
	if err != nil {
		log.Print(err)
		return
	}
	if player.protocol < V1_14 {
		err = player.WriteUInt8(uint8(packet.Difficulty))
		if err != nil {
			log.Print(err)
			return
		}
	}
	if player.protocol >= V1_15 {
		err = player.WriteUInt64(packet.HashedSeed)
		if err != nil {
			log.Print(err)
			return
		}
	}
	err = player.WriteUInt8(packet.MaxPlayers)
	if err != nil {
		log.Print(err)
		return
	}
	err = player.WriteString(string(packet.LevelType))
	if err != nil {
		log.Print(err)
		return
	}
	if player.protocol >= V1_14 {
		err = player.WriteVarInt(32)
		if err != nil {
			log.Print(err)
			return
		}
	}
	if player.protocol > V1_7_6 {
		err = player.WriteBool(packet.ReducedDebug)
		if err != nil {
			log.Print(err)
			return
		}
	}
	if player.protocol >= V1_15 {
		err = player.WriteBool(packet.EnableRespawnScreen)
		if err != nil {
			log.Print(err)
			return
		}
	}
	return
}
func (packet *PacketPlayJoinGame) Handle(player *Player) {}
func (packet *PacketPlayJoinGame) Id() (int, Protocol) {
	return 0x23, V1_10
}

type PacketPlayerPositionLook struct {
	X          float64
	Y          float64
	Z          float64
	Yaw        float32
	Pitch      float32
	Flags      uint8
	TeleportId int
}

func (packet *PacketPlayerPositionLook) Read(player *Player, length int) (err error) {
	return
}
func (packet *PacketPlayerPositionLook) Write(player *Player) (err error) {
	err = player.WriteFloat64(packet.X)
	if err != nil {
		log.Print(err)
		return
	}
	err = player.WriteFloat64(packet.Y)
	if err != nil {
		log.Print(err)
		return
	}
	err = player.WriteFloat64(packet.Z)
	if err != nil {
		log.Print(err)
		return
	}
	err = player.WriteFloat32(packet.Yaw)
	if err != nil {
		log.Print(err)
		return
	}
	err = player.WriteFloat32(packet.Pitch)
	if err != nil {
		log.Print(err)
		return
	}
	err = player.WriteUInt8(packet.Flags)
	if err != nil {
		log.Print(err)
		return
	}
	if player.protocol > V1_8 {
		err = player.WriteVarInt(packet.TeleportId)
		if err != nil {
			log.Print(err)
			return
		}
	}
	return
}
func (packet *PacketPlayerPositionLook) Handle(player *Player) {}
func (packet *PacketPlayerPositionLook) Id() (int, Protocol) {
	return 0x2E, V1_10
}

type PacketUpdateHealth struct {
	Health         float32
	Food           int
	FoodSaturation float32
}

func (packet *PacketUpdateHealth) Read(player *Player, length int) (err error) {
	return
}
func (packet *PacketUpdateHealth) Write(player *Player) (err error) {
	err = player.WriteFloat32(packet.Health)
	if err != nil {
		log.Print(err)
		return
	}
	err = player.WriteVarInt(packet.Food)
	if err != nil {
		log.Print(err)
		return
	}
	err = player.WriteFloat32(packet.FoodSaturation)
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketUpdateHealth) Handle(player *Player) {}
func (packet *PacketUpdateHealth) Id() (int, Protocol) {
	return 0x3E, V1_10
}

type PacketPlayerListHeaderFooter struct {
	Header *string
	Footer *string
}

func (packet *PacketPlayerListHeaderFooter) Read(player *Player, length int) (err error) {
	return
}
func (packet *PacketPlayerListHeaderFooter) Write(player *Player) (err error) {
	var str string
	if packet.Header == nil {
		str = `{"translate":""}`
	} else {
		str = *packet.Header
	}
	err = player.WriteString(str)
	if err != nil {
		log.Print(err)
		return
	}
	if packet.Footer == nil {
		str = `{"translate":""}`
	} else {
		str = *packet.Footer
	}
	err = player.WriteString(str)
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketPlayerListHeaderFooter) Handle(player *Player) {}
func (packet *PacketPlayerListHeaderFooter) Id() (int, Protocol) {
	return 0x47, V1_10
}
