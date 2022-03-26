package main

import(
	"log"
    "bytes"
	"io/ioutil"
	"net/http"
    "encoding/json"
    "regexp"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)


const TG_API_TOKEN = ""
const BOT_NAME = ""
const BOT_MSG_ERROR = "Не знаю, что и ответить..."

type AiResp struct{
    Resp string `json:"responses"`
}

type Req struct{
	Inst [1]Cont `json:"instances"`
}

type Cont struct{
	Con [1][]string `json:"contexts"`
}

func main() {

    botName := regexp.MustCompile("bot_name")
    var users = make(map[int64][]string)

    bot, err := tgbotapi.NewBotAPI(TG_API_TOKEN)
    if err != nil {
        log.Panic(err)
    }

    bot.Debug = true

    log.Printf("Authorized on account %s", bot.Self.UserName)

    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60

    updates := bot.GetUpdatesChan(u)

    for update := range updates {
        if update.Message == nil {
            continue
        }

        users[update.Message.Chat.ID] = append(users[update.Message.Chat.ID], update.Message.Text)
        
        aiMsg, err := GetAiResponse(users[update.Message.Chat.ID])

        msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
        if err != nil{
            msg.Text = BOT_MSG_ERROR
        } else {
            bi := botName.FindStringIndex(aiMsg)
            if len(bi) > 0{
                aiMsg = aiMsg[0:bi[0]-1] + BOT_NAME + aiMsg[bi[1]:]
            }
            msg.Text = aiMsg
            users[update.Message.Chat.ID] = append(users[update.Message.Chat.ID], aiMsg)
        }

        if _, err := bot.Send(msg); err != nil {
            log.Print(err)
        }
    }

}

func GetAiResponse(context []string)(string, error){

	reqBody, err := CreateReqBodyForAi(context)
	resp, err := http.Post("https://api.aicloud.sbercloud.ru/public/v2/boltalka/predict", "application/json", reqBody)
	if err != nil {
	   return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
        return "", err
	}

    var mResp AiResp
	err = json.Unmarshal(body, &mResp)
	if err != nil {
		return "", err
	}
    return mResp.Resp[2:len(mResp.Resp)-2], nil

}

func CreateReqBodyForAi(context []string) (*bytes.Buffer, error){
    /*
    Понятия не имею, как оно там пакуется, я потратил более часа, 
    перебирая структуры, чтобы подготовить JSON для ии сбера,
    а главное, не понятно, почему нельзя было сделать JSON по проще,
    если надо всего-то скинуть массив с фразами.
    */
	  c := Cont{[1][]string{context}}
	  r := Req{[1]Cont{c}}

	  postBody, err := json.Marshal(r)
	  if err != nil {
		return nil, err
	  }

	return bytes.NewBuffer(postBody), nil
}