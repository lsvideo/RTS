package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	//USERNAME = "lishuo"
	//PASSWORD = "123456"
	NETWORK = "tcp"
	//	SERVER   = "127.0.0.1"
	//PORT     = 3306
	//DATABASE = "Poc_star_en"
)

var mysqldb sql.DB

func initMysql(SERVER string, PORT int, USERNAME string, PASSWORD string, DATABASE string) {
	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s", USERNAME, PASSWORD, NETWORK, SERVER, PORT, DATABASE)
	log.Warningln("open DB: ", dsn)
	db, err := sql.Open("mysql", dsn)
	mysqldb = *db
	if err != nil {
		fmt.Printf("Open mysql failed,err:%v\n", err)
		return
	}
	mysqldb.SetConnMaxLifetime(10 * time.Second) //最大连接周期，超过时间的连接就close
	mysqldb.SetMaxOpenConns(100)                 //设置最大连接数
	mysqldb.SetMaxIdleConns(16)                  //设置闲置连接数

	log.Warningln("open DB: ", mysqldb)
}

func closeMysql() {
	mysqldb.Close()
}

/*
+------------+--------------+------+-----+---------+----------------+
| video_id   | int(11)      | NO   | PRI | NULL    | auto_increment |
| user_id    | int(11)      | NO   |     | NULL    |                |
| company_id | int(11)      | NO   |     | NULL    |                |
| server     | varchar(50)  | NO   |     | NULL    |                |
| file_name  | varchar(100) | NO   |     | NULL    |                |
| start_time | bigint(20)   | NO   |     | NULL    |                |
| end_time   | bigint(20)   | NO   |     | NULL    |                |
| size       | int(11)      | NO   |     | NULL    |                |
| status     | tinyint(2)   | NO   |     | NULL    |                |
| type       | int(11)      | NO   |     | NULL    |                |
| sid        | varchar(50)  | NO   |     | NULL    |                |
| to_id      | int(11)      | NO   |     | 0       |                |
| resolution | smallint(6)  | YES  |     | 360     |                |
+------------+--------------+------+-----+---------+----------------+
*/
func insertDvrItem(user *srs_eChatUser, dvr *eChatDvr) {
	result, err := mysqldb.Exec("insert INTO tb_video_file(user_id, company_id, server, file_name, start_time, end_time, size, status, type, sid, to_id, resolution)"+
		"values(?,?,?,?,?,?,?,?,?,?,?,?)", user.User.Uid, user.User.Cid, dvr.Url, dvr.Name, dvr.Start, dvr.Start+dvr.Duration, dvr.Size, dvr.Status, user.User.Action, user.User.Sid, user.User.Toid, user.User.Resolution)
	if err != nil {
		log.Errorln("Insert failed, err: ", err)
		return
	}
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		log.Errorln("Get lastInsertID failed,err", err)
		return
	}
	log.Debugln("LastInsertID:", lastInsertID)
	rowsaffected, err := result.RowsAffected()
	if err != nil {
		log.Errorln("Get RowsAffected failed,err: ", err)
		return
	}
	log.Debugln("RowsAffected:", rowsaffected)
}

//更新数据
func updateDvrStatus(id int, status DvrStatus) {
	result, err := mysqldb.Exec("UPDATE tb_video_file set status=? where video_id=?", status, id)
	if err != nil {
		log.Errorln("Insert failed,err: ", err)
		return
	}
	rowsaffected, err := result.RowsAffected()
	if err != nil {
		log.Errorln("Get RowsAffected failed,err: ", err)
		return
	}
	log.Debugln("RowsAffected:", rowsaffected)
}

//删除数据
func deleteDvrItem(id int) {
	result, err := mysqldb.Exec("delete from tb_video_file where video_id=?", id)
	if err != nil {
		log.Errorln("Insert failed,err: ", err)
		return
	}
	rowsaffected, err := result.RowsAffected()
	if err != nil {
		log.Errorln("Get RowsAffected failed,err: ", err)
		return
	}
	log.Debugln("RowsAffected:", rowsaffected)
}

func queryDvrIdNamebyStatus(status DvrStatus, server string, mapName map[int]string) {
	var id int
	var name string
	rows, err := mysqldb.Query("select video_id, file_name from tb_video_file where status = ? and server = ?", status, server)
	defer func() {
		if rows != nil {
			rows.Close() //可以关闭掉未scan连接一直占用
		}
	}()
	if err != nil {
		log.Errorln("Query failed,err: ", err)
		return
	}
	for rows.Next() {
		err = rows.Scan(&id, &name) //不scan会导致连接不释放
		if err != nil {
			log.Errorln("Scan failed,err ", err)
			return
		}
		mapName[id] = name
		log.Debugln(status, ": ", id, " ", name)
	}
	return
}

func queryDvrIdbyName(name string, server string, mapName map[int]string) {
	var id int
	rows, err := mysqldb.Query("select video_id from tb_video_file where file_name = ? and server = ?", name, server)
	defer func() {
		if rows != nil {
			rows.Close() //可以关闭掉未scan连接一直占用
		}
	}()
	if err != nil {
		log.Errorln("Query failed,err: ", err)
		return
	}
	for rows.Next() {
		err = rows.Scan(&id) //不scan会导致连接不释放
		if err != nil {
			log.Errorln("Scan failed,err ", err)
			return
		}
		mapName[id] = name
		log.Debugln(name, ": ", id)
	}
	return
}
