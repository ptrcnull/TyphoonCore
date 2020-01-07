package main

import (
	t "github.com/TyphoonMC/TyphoonCore"
)

func main() {
	core := t.Init()
	core.SetBrand("Limbo")

	//loadConfig(core)

	core.On(func(e *t.PlayerJoinEvent) {
		//if config.JoinMessage != nil {
		//	e.Player.SendRawMessage(string(config.JoinMessage))
		//}
		//if &playerListHF != nil {
		//	e.Player.WritePacket(&playerListHF)
		//}
		msg := t.ChatMessage("")
		msg.SetExtra([]t.IChatComponent{
			t.ChatMessage("Witaj w poczekalni! Poczekaj na połączenie z serwerem."),
		})
		e.Player.SendMessage(msg)
	})

	core.On(func(e *t.PlayerChatEvent) {
		msg := t.ChatMessage("")
		msg.SetExtra([]t.IChatComponent{
			t.ChatMessage("<"),
			t.ChatMessage(e.Player.GetName()),
			t.ChatMessage("> "),
			t.ChatMessage(e.Message),
		})
		core.GetPlayerRegistry().ForEachPlayer(func(player *t.Player) {
			player.SendMessage(msg)
		})
	})

	core.Start()
}
