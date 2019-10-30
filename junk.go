package main

// func main() {
// 	recRawMsg := []byte(`{"name":"channel add",` +
// 		`"data":{"name":"Hardware support"}}`)
// 	var recMessage Message
// 	err := json.Unmarshal(recRawMsg, &recMessage)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	//fmt.Printf("%#v\n", recMessage)

// 	if recMessage.Name == "channel add" {
// 		channel, err := addChannel(recMessage.Data)
// 		var sendMessage Message
// 		sendMessage.Name = "channel add"
// 		sendMessage.Data = channel
// 		sendRawMessage, err := json.Marshal(sendMessage)
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		fmt.Printf(string(sendRawMessage))
// 	}
// }

// func addChannel(data interface{}) (Channel, error) {
// 	var channel Channel
// 	channelMap := data.(map[string]interface{})
// 	channel.Name = channelMap["name"].(string)
// 	channel.ID = "1"
// 	return channel, nil
// }
