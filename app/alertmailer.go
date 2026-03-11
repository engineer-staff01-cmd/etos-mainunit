package app

import (
	"bytes"
	"embed"
	"fmt"

	//	"io/ioutil"

	"sync"
	"text/template"
	"time"
)

// go:embed ../public/*
//var fs embed.FS

type AlertMail interface {
	// センサー監視アラートメール送信
	SendSensorAlertMail(s *SensorMonitorMailAlert, addresses []UsersMailAddress, apiKey, baseName string) error
	// 通信接続アラートメール送信
	SendCommunicationAlertMail(c *CommunicationMailAlert, addresses []UsersMailAddress, apiKey, baseName string) error
	// バッテリーレベルのアラートメール送信
	SendBatteryLevelAlertMail(b *BatteryLevelMailAlert, addresses []UsersMailAddress, apiKey, baseName string) error
	// デマンド監視アラートメール送信
	SendDemandStatusAlertMail(d *DemandStatusMailAlert, addresses []UsersMailAddress, apiKey, baseName string) error
	// 社屋停電アラートメール送信
	SendPowerOutageAlertMail(u *UpsMailAlert, addresses []UsersMailAddress, apiKey, baseName string) error
}

const (
	subjectTmplText = "[ecoRAMDAR Alert{{.EnvName}}] 拠点名：{{.BaseName}}　アラートタイプ：{{.AlertType}}"

	alertTypeSensor        = "センサー監視"
	alertTypeCommunication = "通信アラート"
	alertTypeBatteryLevel  = "電池残量"
	alertTypeDemand        = "デマンド"
	alertTypeUps           = "親機電源"
)

var sensorCategoryNames = map[string]string{
	strEnergySensor:        strCategoryEnergySensor,
	strEnvironmentalSensor: strCategoryEnvironmentalSensor,
	strInputContact:        strCategoryInputContact,
	strOutputContact:       strCategoryOutputContact,
	strDemandPulseUnit:     strCategoryDemandPulseUnit,
	strChildUnit:           strCategoryChildUnit,
}

var alertStatusNames = map[string]string{
	StrExcess:     StrStatusExcess,
	Strlimit:      StrStatuslimit,
	StrBeVigilant: StrStatusBeVigilant,
	StrRecover:    StrStatusRecover,
	StrNormal:     StrStatusNormal,
}

const undefinedCategory = "未定義"

var (
	dateFormat = "2006/01/02 15:04:05"

	sensorTemplateFilepath        = "public/mail/sensor_alert.go.tpl"
	communicationTemplateFilepath = "public/mail/communication_alert.go.tpl"
	batteryTemplateFilepath       = "public/mail/battery_alert.go.tpl"
	demandTemplateFilepath        = "public/mail/demand_alert.go.tpl"
	upsTemplateFilepath           = "public/mail/ups_alert.go.tpl"
)

type AlertMailer struct {
	mailer     *SendGridMailer
	subjectTpl *template.Template
	sensorTpl  *template.Template
	commTpl    *template.Template
	battTpl    *template.Template
	demandTpl  *template.Template
	upsTpl     *template.Template
	env        Env
	envName    string
}

func NewAlertMailer(env Env, fs_public embed.FS) (*AlertMailer, error) {
	Logger.Writef(LOG_LEVEL_DEBUG, "Parse(subjectTmplText)")
	subjectTpl, err := template.New("").Parse(subjectTmplText)
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "Failed to Parse(subjectTmplText) %+v", err.Error())
		return nil, fmt.Errorf("Failed to parse subjectTpl AlertMailer: %w", err)
	}

	Logger.Writef(LOG_LEVEL_DEBUG, "loadTemplate(sensorTemplateFilepath) :%+v", &fs_public)
	sensorTpl, err := loadTemplate(sensorTemplateFilepath, fs_public)
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "Failed to loadTemplate(sensorTemplateFilepath) %+v", err.Error())
		return nil, fmt.Errorf("Failed to parse sensorTpl AlertMailer: %w", err)
	}

	Logger.Writef(LOG_LEVEL_DEBUG, "loadTemplate(communicationTemplateFilepath)")
	commTpl, err := loadTemplate(communicationTemplateFilepath, fs_public)
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "Failed to loadTemplate(communicationTemplateFilepath) %+v", err.Error())
		return nil, fmt.Errorf("Failed to parse commTpl AlertMailer: %w", err)
	}

	Logger.Writef(LOG_LEVEL_DEBUG, "loadTemplate(batteryTemplateFilepath)")
	battTpl, err := loadTemplate(batteryTemplateFilepath, fs_public)
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "Failed to loadTemplate(batteryTemplateFilepath) %+v", err.Error())
		return nil, fmt.Errorf("Failed to parse battTpl AlertMailer: %w", err)
	}

	Logger.Writef(LOG_LEVEL_DEBUG, "loadTemplate(demandTemplateFilepath)")
	demandTpl, err := loadTemplate(demandTemplateFilepath, fs_public)
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "Failed to loadTemplate(demandTemplateFilepath) %+v", err.Error())
		return nil, fmt.Errorf("Failed to parse demandTpl AlertMailer: %w", err)
	}

	Logger.Writef(LOG_LEVEL_DEBUG, "loadTemplate(upsTemplateFilepath)")
	upsTpl, err := loadTemplate(upsTemplateFilepath, fs_public)
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "Failed to loadTemplate(upsTemplateFilepath) %+v", err.Error())
		return nil, fmt.Errorf("Failed to parse upsTpl AlertMailer: %w", err)
	}

	var envName = ""
	if env.Name != Production {
		envName = " " + env.Name
	}

	return &AlertMailer{
		mailer:     NewSendGridMailer(),
		subjectTpl: subjectTpl,
		sensorTpl:  sensorTpl,
		commTpl:    commTpl,
		battTpl:    battTpl,
		demandTpl:  demandTpl,
		upsTpl:     upsTpl,
		env:        env,
		envName:    envName,
	}, nil
}

func loadTemplate(filepath string, fs_public embed.FS) (*template.Template, error) {
	/*
		f, err := fs_public.Open(filepath)
		if err != nil {
			return nil, err
		}

		defer f.Close()
	*/
	//	contents, err := ioutil.ReadAll(f)
	contents, err := fs_public.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	tpl, err := template.New(filepath).Parse(string(contents))
	if err != nil {
		return nil, fmt.Errorf("Failed to parse %s file: %w", filepath, err)
	}
	return tpl, nil
}

func generateText(tmpl *template.Template, input map[string]interface{}) (string, error) {
	var body bytes.Buffer
	if err := tmpl.Execute(&body, input); err != nil {
		return "", err
	}
	return body.String(), nil
}

func selectSensorNameBySensorKind(sensorKind string) string {
	for kind, category := range sensorCategoryNames {
		if kind == sensorKind {
			return category
		}
	}
	return undefinedCategory
}

func selectAlertStatusNameByStatus(statusEng string) string {
	for eng, jp := range alertStatusNames {
		if eng == statusEng {
			return jp
		}
	}
	return undefinedCategory
}

// SendSensorAlertMail センサー監視アラートメール送信
func (am *AlertMailer) SendSensorAlertMail(s *SensorMonitorMailAlert, addresses []UsersMailAddress, apiKey, baseName string) error {

	subjectInput := map[string]interface{}{
		"EnvName":   am.envName,
		"BaseName":  baseName,
		"AlertType": alertTypeSensor,
	}

	bodyInput := map[string]interface{}{
		// 2022-08-25 本文内の発生日時をメール送信日時に合わせる対応を行った
		// "Occured":        time.Unix(s.OccurredAt, 0).Format(dateFormat),
		"Occured":        time.Now().Format(dateFormat),
		"SensorCategory": selectSensorNameBySensorKind(s.Kind),
		"AlertMessage":   s.AlertMessage,
		"URL":            am.env.AdminHost,
	}

	subject, err := generateText(am.subjectTpl, subjectInput)
	if err != nil {
		return err
	}

	body, err := generateText(am.sensorTpl, bodyInput)
	if err != nil {
		return err
	}

	for _, to := range addresses {
		if to.EnableSensorAlert != 0 {
			if err := am.mailer.Send(apiKey, to.Email, subject, body); err != nil {
				// メールを送信できなかった場合の代替手段はないため、調査用にエラーログを出力するのみにしておく
				Logger.Writef(LOG_LEVEL_ERR, "SendSensorAlertMail to:%s error: %s", to.Email, err.Error())
			}
		}
	}

	return nil
}

// SendCommunicationAlertMail 通信アラートメール送信
func (am *AlertMailer) SendCommunicationAlertMail(c *CommunicationMailAlert, addresses []UsersMailAddress, apiKey, baseName string) error {

	subjectInput := map[string]interface{}{
		"EnvName":   am.envName,
		"BaseName":  baseName,
		"AlertType": alertTypeCommunication,
	}

	bodyInput := map[string]interface{}{
		// 2022-08-25 本文内の発生日時をメール送信日時に合わせる対応を行った
		// "Occured":        time.Unix(s.OccurredAt, 0).Format(dateFormat),
		"Occured":    time.Now().Format(dateFormat),
		"SensorID":   c.SensorID,
		"SensorName": c.SensorName,
		"URL":        am.env.AdminHost,
	}

	subject, err := generateText(am.subjectTpl, subjectInput)
	if err != nil {
		return err
	}

	body, err := generateText(am.commTpl, bodyInput)
	if err != nil {
		return err
	}

	for _, to := range addresses {
		if to.EnableConnectionAlert != 0 {
			if err := am.mailer.Send(apiKey, to.Email, subject, body); err != nil {
				// メールを送信できなかった場合の代替手段はないため、調査用にエラーログを出力するのみにしておく
				Logger.Writef(LOG_LEVEL_ERR, "SendCommunicationAlertMail to:%s error: %s", to.Email, err.Error())
			}
		}
	}

	return nil
}

// SendBatteryLevelAlertMail バッテリーレベルのアラートメール送信
func (am *AlertMailer) SendBatteryLevelAlertMail(b *BatteryLevelMailAlert, addresses []UsersMailAddress, apiKey, baseName string) error {

	subjectInput := map[string]interface{}{
		"EnvName":   am.envName,
		"BaseName":  baseName,
		"AlertType": alertTypeBatteryLevel,
	}

	bodyInput := map[string]interface{}{
		// 2022-08-25 本文内の発生日時をメール送信日時に合わせる対応を行った
		// "Occured":        time.Unix(s.OccurredAt, 0).Format(dateFormat),
		"Occured": time.Now().Format(dateFormat),
		"URL":     am.env.AdminHost,
	}

	subject, err := generateText(am.subjectTpl, subjectInput)
	if err != nil {
		return err
	}

	body, err := generateText(am.battTpl, bodyInput)
	if err != nil {
		return err
	}

	for _, to := range addresses {
		if to.EnableBatteryAlert != 0 {
			if err := am.mailer.Send(apiKey, to.Email, subject, body); err != nil {
				// メールを送信できなかった場合の代替手段はないため、調査用にエラーログを出力するのみにしておく
				Logger.Writef(LOG_LEVEL_ERR, "SendBatteryLevelAlertMail to:%s error: %s", to.Email, err.Error())
			}
		}
	}

	return nil
}

// SendDemandStatusAlertMail デマンド監視アラートメール送信
func (am *AlertMailer) SendDemandStatusAlertMail(d *DemandStatusMailAlert, addresses []UsersMailAddress, apiKey, baseName string) error {
	subjectInput := map[string]interface{}{
		"EnvName":   am.envName,
		"BaseName":  baseName,
		"AlertType": alertTypeDemand,
	}

	round := func(n float64) string {
		return fmt.Sprintf("%.0f", n)
	}

	bodyInput := map[string]interface{}{
		// 2022-08-25 本文内の発生日時をメール送信日時に合わせる対応を行った
		// "Occured":        time.Unix(s.OccurredAt, 0).Format(dateFormat),
		"Occured":                time.Now().Format(dateFormat),
		"AlertStatus":            selectAlertStatusNameByStatus(d.AlarmStatus),
		"CurrentElectricPower":   round(d.CurrentElectricPower),
		"PredictedElectricPower": round(d.PredictedElectricPower),
		"AdjustedElectricPower":  round(d.AdjustedElectricPower),
		"TargetElectricPower":    round(d.TargetElectricPower),
		"LimitElectricPower":     round(d.LimitElectricPower),
		"ContractElectricPower":  round(d.ContractElectricPower),
		"URL":                    am.env.AdminHost,
	}

	subject, err := generateText(am.subjectTpl, subjectInput)
	if err != nil {
		return err
	}

	body, err := generateText(am.demandTpl, bodyInput)
	if err != nil {
		return err
	}

	for _, to := range addresses {
		if to.EnableDemandAlert != 0 {
			if err := am.mailer.Send(apiKey, to.Email, subject, body); err != nil {
				// メールを送信できなかった場合の代替手段はないため、調査用にエラーログを出力するのみにしておく
				Logger.Writef(LOG_LEVEL_ERR, "SendDemandStatusAlertMail to:%s error: %s", to.Email, err.Error())
			}
		}
	}

	return nil
}

// SendPowerOutageAlertMail 社屋停電アラートメール送信
func (am *AlertMailer) SendPowerOutageAlertMail(u *UpsMailAlert, addresses []UsersMailAddress, apiKey, baseName string) error {
	subjectInput := map[string]interface{}{
		"EnvName":   am.envName,
		"BaseName":  baseName,
		"AlertType": alertTypeUps,
	}

	bodyInput := map[string]interface{}{
		// 2022-08-25 本文内の発生日時をメール送信日時に合わせる対応を行った
		// "Occured":        time.Unix(s.OccurredAt, 0).Format(dateFormat),
		"Occured": time.Now().Format(dateFormat),
		"URL":     am.env.AdminHost,
	}

	subject, err := generateText(am.subjectTpl, subjectInput)
	if err != nil {
		return err
	}

	body, err := generateText(am.upsTpl, bodyInput)
	if err != nil {
		return err
	}

	for _, to := range addresses {
		if to.EnableUpsAlert != 0 {
			if err := am.mailer.Send(apiKey, to.Email, subject, body); err != nil {
				// メールを送信できなかった場合の代替手段はないため、調査用にエラーログを出力するのみにしておく
				Logger.Writef(LOG_LEVEL_ERR, "SendPowerOutageAlertMail to:%s error: %s", to.Email, err.Error())
			}
		}
	}

	return nil
}

type MailThread struct {
	to     chan ChannelMessage
	mailer *AlertMailer
}

func NewMailThread(env Env, fs_public embed.FS) (*MailThread, error) {
	mailer, err := NewAlertMailer(env, fs_public)
	if err != nil {
		return nil, fmt.Errorf("Failed spawning a mail thread")
	}

	return &MailThread{
		to:     make(chan ChannelMessage, MailChannelBufferSize), // メール送信専用(watcher -> mailer)
		mailer: mailer,
	}, nil
}

func (mt *MailThread) Run(wg *sync.WaitGroup) {
	Logger.Writef(LOG_LEVEL_DEBUG, "Start MailThread")
	wg.Add(1) // goroutine起動前にカウントを増やす
	go func() {
		mt.childThreadRun()
		Logger.Writef(LOG_LEVEL_DEBUG, "Stop MailThread")
		wg.Done()
	}()
}

func (mt *MailThread) childThreadRun() {
	for {
		chmsg := <-mt.to
		switch chmsg.messageType {
		case End:
			return

		case Alert:
			base := BaseInformation{}
			if err := database.SelectOne(&base); err != nil {
				Logger.Writef(LOG_LEVEL_WARNING, "Failed to select base information: %s", err.Error())
				continue
			}

			addresses := []UsersMailAddress{}
			if err := database.SelectAll(&addresses); err != nil {
				Logger.Writef(LOG_LEVEL_WARNING, "Failed to select user mail addresses: %s", err.Error())
				continue
			}
			if addresses == nil {
				return
			}

			// センサーアラートに関するメールの送信
			mt.sendSensorMonitorAlertMails(addresses, base.SendGridAPIKey, base.Name)
			// 通信アラートに関するメールの送信
			mt.sendCommunicationAlertMails(addresses, base.SendGridAPIKey, base.Name)
			// 電池残量に関するアラートメール送信
			mt.sendBatteryLevelAlertMails(addresses, base.SendGridAPIKey, base.Name)
			// デマンド監視に関するアラートメール送信
			mt.sendDemandStatusAlertMails(addresses, base.SendGridAPIKey, base.Name)
			// 社屋停電に関するアラートメール送信
			//if RebootCommandFlag == false {
			mt.sendUpsAlertMails(addresses, base.SendGridAPIKey, base.Name)
			//}
		}
	}
}

func (mt *MailThread) sendSensorMonitorAlertMails(addresses []UsersMailAddress, apiKey, baseName string) {
	sensorMonitorMailAlerts := []SensorMonitorMailAlert{}
	if err := database.SelectAll(&sensorMonitorMailAlerts); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select sensor monitor mail alerts: %s", err.Error())
		return
	}
	for _, s := range sensorMonitorMailAlerts {
		Logger.Writef(LOG_LEVEL_INFO, "sendSensorMonitorAlertMails to:%v", addresses)
		if err := mt.mailer.SendSensorAlertMail(&s, addresses, apiKey, baseName); err != nil {
			// メールを送信できなかった場合の代替手段はないため、調査用にエラーログを出力するのみにしておく
			Logger.Writef(LOG_LEVEL_ERR, "sendSensorMonitorAlertMails:%s", err.Error())
			continue
		}
		database.Delete(&s, "time = ? AND sensor_id = ?", s.Time, s.SensorID)
	}
}

func (mt *MailThread) sendCommunicationAlertMails(addresses []UsersMailAddress, apiKey, baseName string) {
	communicationMailAlerts := []CommunicationMailAlert{}
	if err := database.SelectAll(&communicationMailAlerts); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select communication mail alerts: %s", err.Error())
		return
	}
	for _, c := range communicationMailAlerts {
		Logger.Writef(LOG_LEVEL_INFO, "sendCommunicationAlertMails to:%v", addresses)
		if err := mt.mailer.SendCommunicationAlertMail(&c, addresses, apiKey, baseName); err != nil {
			// メールを送信できなかった場合の代替手段はないため、調査用にエラーログを出力するのみにしておく
			Logger.Writef(LOG_LEVEL_ERR, "sendCommunicationAlertMails:%s", err.Error())
			continue
		}
		database.Delete(&c, "time = ? AND sensor_id = ?", c.Time, c.SensorID)
	}
}

func (mt *MailThread) sendBatteryLevelAlertMails(addresses []UsersMailAddress, apiKey, baseName string) {
	batteryLevelMailAlerts := []BatteryLevelMailAlert{}
	if err := database.SelectAll(&batteryLevelMailAlerts); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select battery level mail alerts: %s", err.Error())
		return
	}
	for _, b := range batteryLevelMailAlerts {
		Logger.Writef(LOG_LEVEL_INFO, "sendBatteryLevelAlertMails to:%v", addresses)
		if err := mt.mailer.SendBatteryLevelAlertMail(&b, addresses, apiKey, baseName); err != nil {
			// メールを送信できなかった場合の代替手段はないため、調査用にエラーログを出力するのみにしておく
			Logger.Writef(LOG_LEVEL_ERR, "sendBatteryLevelAlertMails:%s", err.Error())
			continue
		}
		database.Delete(&b, "time = ? AND sensor_id = ?", b.Time, b.SensorID)
	}
}

func (mt *MailThread) sendDemandStatusAlertMails(addresses []UsersMailAddress, apiKey, baseName string) {
	demandStatusMailAlerts := []DemandStatusMailAlert{}
	if err := database.SelectAll(&demandStatusMailAlerts); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select demand status mail alerts: %s", err.Error())
		return
	}
	for _, d := range demandStatusMailAlerts {
		Logger.Writef(LOG_LEVEL_INFO, "sendDemandStatusAlertMails to:%v", addresses)
		if err := mt.mailer.SendDemandStatusAlertMail(&d, addresses, apiKey, baseName); err != nil {
			// メールを送信できなかった場合の代替手段はないため、調査用にエラーログを出力するのみにしておく
			Logger.Writef(LOG_LEVEL_ERR, "sendDemandStatusAlertMails:%s", err.Error())
			continue
		}
		database.Delete(&d, "time = ?", d.Time)
	}
}

func (mt *MailThread) sendUpsAlertMails(addresses []UsersMailAddress, apiKey, baseName string) {
	upsMailAlerts := []UpsMailAlert{}
	if err := database.SelectAll(&upsMailAlerts); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select UPS mail alerts: %s", err.Error())
		return
	}
	for _, u := range upsMailAlerts {
		Logger.Writef(LOG_LEVEL_INFO, "sendUpsAlertMails to:%v", addresses)
		if err := mt.mailer.SendPowerOutageAlertMail(&u, addresses, apiKey, baseName); err != nil {
			// メールを送信できなかった場合の代替手段はないため、調査用にエラーログを出力するのみにしておく
			Logger.Writef(LOG_LEVEL_ERR, "sendUpsAlertMails:%s", err.Error())
			continue
		}
		database.Delete(&u, "time = ?", u.Time)
	}
}
