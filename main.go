package main

import (
	"net/http"
	"io/ioutil"
	"fmt"
	"time"
	"strconv"
	"log"
	_"github.com/max75025/go-sqlite3"
	"encoding/json"
	"os"
	"database/sql"
	"strings"
)
const constEndTime = 9999999999999999

const testApiKey  = "5a9ebd7d5f7c8cc17f385f2b36b26181a03fb3dfe78c512cb71f538869a7ea8d6b803385245dfcb698d47be097c82d4759eed12ad106021e2cfa646f905cacfc"
const testApiStartTime = 1532449279
const monthInSecond = 2592000


var newEvent = false
var newAV = false


type event struct{
	DateTime 	int
	TypeTrace 	[]string
	ResultTypes []string
	IpAddr		string
	Country		string
}

type av struct{
	ApiKey					string
	EventTime				int
	EventType				string
	FileName				string
	FileExt					string
	FilePath				string
	SuspiciousType			string
	SuspiciousDescripton 	string
}

func saveAVToDB(db *sql.DB, jsonStr string) error{
	if jsonStr!="null"{


		var av []av
		json.Unmarshal([]byte(jsonStr),&av)
		for _,k:= range av{
			stmt,err:= db.Prepare("INSERT INTO AV(ApiKey, EventTime, EventType, FileName, FileExt, FilePath, SuspiciousType, SuspiciousDescripton) VALUES (?,?,?,?,?,?,?,?)")
			if err!= nil{
				log.Println(err)
				return err
			}
			_, err = stmt.Exec(k.ApiKey, k.EventTime, k.EventType, k.FileName, k.FileExt, k.FilePath, k.SuspiciousType, k.SuspiciousDescripton)
			if err!=nil{
				log.Println(err)
				return err
			}
		}
		newAV = true
	}

	return nil
}

func saveEventToDB(db *sql.DB, jsonStr string) error{
	if jsonStr!="null"{

		var events []event
		json.Unmarshal([]byte(jsonStr),&events)
		for _,k:= range events{
			stmt,err:= db.Prepare("INSERT INTO event(DateTime, TypeTrace, ResultTypes, IpAddr, Country) VALUES (?,?,?,?,?)")
			if err!= nil{
				log.Println(err)
				return err
			}
			_, err = stmt.Exec(k.DateTime, strings.Join(k.TypeTrace, ", ") , strings.Join(k.ResultTypes, ", "), k.IpAddr, k.Country)
			if err!=nil{
				log.Println(err)
				return err
			}
		}
		newEvent = true
	}
	return nil
}


func getEventClient(apiKey string, startTime int, endTime int)(string,error)  {
	url := "http://wafwaf.tech/eventclient/" + apiKey + "/" + strconv.Itoa(startTime)+"/"+ strconv.Itoa(endTime)
	//fmt.Println(url)
	resp,err:= http.Get(url)
	if err!= nil {
		//log.Println(err)
		return "",err
	}
	defer resp.Body.Close()
	content,err:= ioutil.ReadAll(resp.Body)
	if err!=nil{
		//log.Println(err)
		return "",err
	}

	return string(content), nil
}

func getAVClient(apiKey string, startTime int, endTime int)(string,error)  {
	url := "http://wafwaf.tech/eventav/" + apiKey + "/" + strconv.Itoa(startTime)+"/"+ strconv.Itoa(endTime)
	resp,err:= http.Get(url)
	if err!= nil {
		return "",err
	}
	defer resp.Body.Close()
	content,err:= ioutil.ReadAll(resp.Body)
	if err!=nil{
		return "",err
	}

	return string(content), nil
}

func autoCheckNewEventAndAvClient(db *sql.DB,apiKey string){


		currentTime := int(time.Now().Unix())
		//fmt.Println("check...")
		result,err:= getEventClient(apiKey,currentTime-10,constEndTime )
		if err!=nil{
			log.Println(err)
		}else{
			saveEventToDB(db, result)
		}

		resultAV,err:= getAVClient(apiKey,currentTime-10,constEndTime )
		if err!=nil{
			log.Println(err)
		}else{
			saveAVToDB(db, result)
		}


		fmt.Println("event: "+result)
		fmt.Println("AV: "+resultAV)
	}
}


func haveNewEvent()bool{
	if newEvent{
		newEvent = false
		return !newEvent
	}
	return newEvent
}

func haveNewAV()bool{
	if newAV{
		newAV = false
		return !newAV
	}
	return newAV
}


func Start(apiKey string){
	startTimeEvent:= int(time.Now().Unix())-monthInSecond
	startTimeAV := startTimeEvent
	newDB:= false

	if _, err := os.Stat("./db.db"); os.IsNotExist(err) {
		_,fileErr:=os.Create("./db.db")
		if fileErr!=nil{log.Println(err)}else{newDB = true}
	}

	db,err:= sql.Open("sqlite3","./db.db" )
	if err!=nil{log.Println(err)}

	if newDB{
		_,err:= db.Exec("CREATE TABLE `event`( `DateTime` INTEGER, `TypeTrace` TEXT , `ResultTypes` TEXT,`IpAddr`	TEXT,`Country` TEXT)")
		if err!=nil{log.Println(err)}
		_,err = db.Exec("CREATE TABLE `AV`( `ApiKey` TEXT, `EventTime` INTEGER , `EventType` TEXT, `FileName`	TEXT,`FileExt` TEXT, `FilePath` TEXT,`SuspiciousType` TEXT, `SuspiciousDescripton` TEXT )")
		if err!=nil{log.Println(err)}
	}else{
		lastTimeEvent:=0
		lastTimeAV :=0
		err = db.QueryRow("SELECT MAX(DateTime) FROM event ").Scan(&lastTimeEvent)
		if err == nil{
			startTimeEvent = lastTimeEvent + 1
		}else{log.Println(err)}
		err = db.QueryRow("SELECT MAX(EventTime) FROM AV ").Scan(&lastTimeAV)
		if err == nil{
			startTimeAV = lastTimeAV + 1
		}else{log.Println(err)}
	}

	fmt.Println("last event time" +strconv.Itoa(startTimeEvent))
	fmt.Println("last AV time" +strconv.Itoa(startTimeAV))

	result,err:=getEventClient(apiKey,startTimeEvent,constEndTime)
	if err!= nil{
		log.Println(err)
	}else{
		fmt.Println(result)
		err = saveEventToDB(db, result)
		if err!= nil{
			log.Println(err)
		}
	}

	result,err =getAVClient(apiKey,startTimeAV,constEndTime)
	if err!= nil{
		log.Println(err)
	}else{
		fmt.Println(result)
		err = saveAVToDB(db, result)
		if err!= nil{
			log.Println(err)
		}
	}

	autoCheckNewEventAndAvClient(db, apiKey)
	db.Close()
}


func main(){
	log.SetFlags(log.Lshortfile)

	Start(testApiKey)

}