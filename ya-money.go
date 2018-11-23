//YandexMoneyClient
package yamoney

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)
//---- Уведомление Http
type HttpParams struct {
	NotificationType string `form:"notification_type"`	//Для переводов из кошелька - p2p-incoming Для переводов с произвольной карты - card-incoming.
	OperationId      string `form:"operation_id"`		//Идентификатор операции в истории счета получателя.
	Amount           float64 `form:"amount"`				//Сумма операции.
	WithDrawAmount float64 `form:"withdraw_amount"`		//Сумма, которая списана со счета отправителя.
	Currency         string `form:"currency"`			//Код валюты счета пользователя. Всегда 643 (рубль РФ согласно ISO 4217).
	DateTime         string `form:"datetime"`			//Дата и время совершения перевода.
	Sender           string `form:"sender"`				//Для переводов из кошелька - номер счета отправителя. Для переводов с произвольной карты - параметр содержит пустую строку.
	Codepro          bool `form:"codepro"`				//Для переводов из кошелька — перевод защищен кодом протекции. Для переводов с произвольной карты — всегда false.
	Label            string `form:"label"`				//Метка платежа. Если метки у платежа нет, параметр содержит пустую строку.
	Sha1Hash			string `form:"sha1_hash"`			//SHA-1 hash параметров уведомления.
	TestNotification	bool `form:"test_notification"`	//Флаг означает, что уведомление тестовое. По умолчанию параметр отсутствует.
	Unaccepted	bool `form:"unaccepted"`				//Флаг означает, что пользователь не получил перевод. Возможные причины: Перевод заморожен, так как на счете получателя достигнут лимит доступного остатка. Отображается в поле hold ответа метода account-info.Перевод защищен кодом протекции. В этом случае codepro=true.
	Title			 string `form:"title"`				//
	Lastname    string `form:"lastname"`				//ФИО отправителя перевода. Если не запрашивались, параметры содержат пустую строку.
	Firstname   string `form:"firstname"`
	Fathersname string `form:"fathersname"`
	Email       string `form:"email"`					//Адрес электронной почты отправителя перевода. Если email не запрашивался, параметр содержит пустую строку.
	Phone       string `form:"phone"`					//Телефон отправителя перевода. Если телефон не запрашивался, параметр содержит пустую строку.
	City       string `form:"city"`					//Адрес, указанный отправителем перевода для доставки. Если адрес не запрашивался, параметры содержат пустую строку.
	Street       string `form:"street"`
	Building       string `form:"building"`
	Suite       string `form:"suite"`
	Flat       string `form:"flat"`
	Zip       string `form:"zip"`
}

func (yp *HttpParams) CheckSha1(nsecret string) bool {
	str := yp.NotificationType + "&" +
		yp.OperationId + "&" +
		fmt.Sprintf("%.2f", yp.Amount) + "&" +
		yp.Currency + "&" +
		yp.DateTime + "&" +
		yp.Sender + "&" +
		fmt.Sprintf("%v", yp.Codepro) + "&" +
		nsecret + "&" +
		yp.Label

	hasher := sha1.New()
	hasher.Write([]byte(str))
	sha := hex.EncodeToString(hasher.Sum(nil))

	if !strings.EqualFold(sha, yp.Sha1Hash) {
		return false
	}
	return true
}

//------

//Структура для метода account-info
type Account struct {
	Account string `json:"account"`	//Номер счета пользователя.
	Balance float64 `json:"balance"`	//Баланс счета пользователя.
	Currency string `json:"currency"`	//Код валюты счета пользователя. Всегда 643 (рубль РФ по стандарту ISO 4217).
	AccountStatus string `json:"account_status"`	//Статус пользователя.
	AccountType string `json:"account_type"`	//Тип счета пользователя.
	Avatar AccountAvatar `json:"avatar"`	//Ссылка на аватар пользователя. Если аватарка пользователя не установлена, параметр отсутствует.
	BalanceDetails AccountBalanceDetails `json:"balance_details"`	//Расширенная информация о балансе. По умолчанию этот блок отсутствует. Блок появляется если сейчас или когда либо ранее были: зачисления в очереди;задолженности;	блокировки средств.
	CardsLinked []AccountCardLinked `json:"cards_linked"`	//нформация о привязанных банковских картах. Если к счету не привязано ни одной карты, параметр отсутствует. Если к счету привязана хотя бы одна карта, параметр содержит список данных о привязанных картах.
}
type AccountStatus string
var(
	Anonymous = AccountStatus("anonymous")	//анонимный счет
	Named = AccountStatus("named")	//именной счет
	Identified = AccountStatus("identified")	//идентифицированный счет
)
type AccountType string
var(
	Personal = AccountType("personal")	//счет пользователя в Яндекс.Деньгах
	Professional = AccountType("professional")	//профессиональный счет в Яндекс.Деньгах
)

type AccountAvatar struct {
	Url string `json:"url"`	//Ссылка на аватар пользователя.
	Ts string `json:"ts"`	//Timestamp последнего изменения аватарки.
}
type AccountBalanceDetails struct {
	Total float64 `json:"total"`	//Общий баланс счета.
	Available float64 `json:"available"`	//Сумма доступная для расходных операций.
	DepositionPending float64 `json:"deposition_pending"`	//Сумма стоящих в очереди пополнений. Если зачислений в очереди нет, то параметр отсутствует.
	Blocked float64 `json:"blocked"`	//Сумма заблокированных средств по решению исполнительных органов. Если заблокированных средств нет то параметр отсутствует.
	Debt float64 `json:"debt"`	//Сумма задолженности (отрицательного остатка на счете). Если задолженности нет, то параметр отсутствует.
	Hold float64 `json:"hold"`	//Сумма замороженных средств. Если замороженных средств нет, то параметр отсутствует.
}
type AccountCardLinked struct {
	PanFragment string `json:"pan_fragment"`	//Маскированный номер карты.
	Type string `json:"type"`	//Тип карты. Может отсутствовать, если неизвестен. Возможные значения:
}



type Operation struct {
	Error string `json:"error"`					//Код ошибки, присутствует при ошибке выполнения запроса.
	OperationID string `json:"operation_id"` //Идентификатор операции. Значение параметра соответствует либо значению параметра operation_id ответа метода operation-history либо, в случае если запрашивается история счета плательщика, значению поля payment_id ответа метода process-payment.
	Status string `json:"status"`			//Статус платежа (перевода). Может принимать следующие значения:
	Pattern_id string `json:"pattern_id"`	//Идентификатор шаблона, по которому совершен платеж. Присутствует только для платежей.
	Direction string `json:"direction"`  //IN - приход OUT расход
	Amount float64 `json:"amount"`		//Сумма операции
	AmountDue float64 `json:"amount_due"`	//Сумма к получению. Присутствует для исходящих переводов другим пользователям.
	Fee float64 `json:"fee"`			//Сумма комиссии. Присутствует для исходящих переводов другим пользователям.
	DateTime string `json:"datetime"`		//Дата и время совершения операции.
	Title string `json:"title"`				//Краткое описание операции (название магазина или источник пополнения).
	Sender string `json:"sender"`			//Номер счета отправителя перевода. Присутствует для входящих переводов от других пользователей.
	Recipient string `json:"recipient"`		//Идентификатор получателя перевода. Присутствует для исходящих переводов другим пользователям.
	Recipient_type string `json:"recipient_type"`	//Тип идентификатора получателя перевода. Присутствует для исходяших переводов другим пользователям.
	Message string `json:"message"`				//Сообщение получателю перевода. Присутствует для переводов другим пользователям.
	Comment string `json:"comment"`				//Комментарий к переводу или пополнению. Присутствует в истории отправителя перевода или получателя пополнения.
	Codepro bool `json:"codepro"`			//Перевод защищен кодом протекции. Присутствует для переводов другим пользователям.
	ProtectionCode string `json:"protection_code"`	//Код протекции. Присутствует для исходящих переводов, защищённых кодом протекции.
	Expires string `json:"expires"`		//Дата и время истечения срока действия кода протекции. Присутствует для входящих и исходящих переводов (от/другим) пользователям, защищённых кодом протекции.
	AnswerDateTime string `json:"answer_datetime"`	//Дата и время приёма или отмены перевода, защищённого кодом протекции. Присутствует для входящих и исходящих переводов, защищённых кодом протекции. Если перевод еще не принят/не отвергнут получателем - поле отсутствует.
	Label string `json:"label"`		//Метка платежа. Присутствует для входящих и исходящих переводов другим пользователям Яндекс.Денег, у которых был указан параметр label вызова request-payment.
	Details string `json:"details"`	//Детальное описание платежа. Строка произвольного формата, может содержать любые символы и переводы строк.
	Type string `json:"type"`		//Тип операции. Описание возможных типов операций см. в описании метода operation-history
	DigitalGoods digitalGoods `json:"digital_goods"`	//Данные о цифровом товаре (пин-коды и бонусы игр, iTunes, Xbox, etc.) Поле присутствует при успешном платеже в магазины цифровых товаров. Описание формата можно найти здесь.
	Account string `json:"account"`	//Номер счета получателя в сервисе Яндекс.Деньги
	Phone string `json:"phone"`	//Номер привязанного мобильного телефона получателя
	Email string `json:"email"`	//Email получателя перевода
}

//Статус операции
type OperationStatus string
var(
	StatusSuccess = OperationStatus("success")		// платеж завершен успешно;
	StatusRefused  = OperationStatus("refused")		// платеж отвергнут получателем или отменен отправителем;
	StatusInProgress = OperationStatus("in_progress")		// платеж не завершен, перевод не принят получателем или ожидает ввода кода протекции.
)
//Возможные типы операций
type OperationType string
var(
	PaymentShop = OperationType("payment-shop")				//Исходящий платеж в магазин
	OutgoingTransfer = OperationType("outgoing-transfer")	//Исходящий P2P-перевод любого типа
	Deposition = OperationType("deposition")				//Зачисление
	IncomingTransfer = OperationType("incoming-transfer")	//Входящий перевод или перевод до востребования.
	IncomingTransferProtected = OperationType("incoming-transfer-protected")	//Входящий перевод с кодом протекции.
)
type OperationError string
var(
	IllegalParamOperationId = OperationError("illegal_param_operation_id")	//Неверное значение параметра operation_id.
	//все прочие значения... Техническая ошибка, повторите вызов метода позднее.
)

type DigitalGood struct {
	MerchantArticleId string `json:"merchantArticleId"`	//Идентификатор товара в системе продавца. Присутствует только для товаров.
	Serial string `json:"serial"`		//Серийный номер товара (открытая часть пин-кода, кода активации или логин).
	Secret string `json:"secret"`		//Секрет цифрового товара (закрытая часть пин-кода, кода активации, пароль или ссылка на скачивание).
}
type digitalGoods struct {
	Article []DigitalGood  `json:"article"`
	Bonus []DigitalGood `json:"bonus"`
}
//--------------

type ByAmount []Operation
type ByDateTime []Operation

func (s ByAmount) Len() int {
	return len(s)
}

func (s ByAmount) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByAmount) Less(i, j int) bool {
	return s[i].Amount < s[j].Amount
}

func (s ByDateTime) Len() int {
	return len(s)
}

func (s ByDateTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByDateTime) Less(i, j int) bool {
	return len(s[i].DateTime) < len(s[j].DateTime)
}


type ResponseOperations struct {
	Error	   string `json"error"`
	NextRecord string `json:"next_record"`
	Operations []Operation `json:"operations"`
}

type OperationHistoryParams struct {
	from time.Time
	till time.Time
}


type YandexMoneyClient struct {
	token string
	api_url string
}

func ( ya YandexMoneyClient ) Account_info() (account Account, err error) {
	str, err := ya.execute("account-info", nil)
	if err != nil {
		return
	}
	err = json.Unmarshal(str, &account)
	if err != nil {
		return
	}
	return
}

func ( ya YandexMoneyClient ) OperationDetails( operation_id string ) (ops Operation, err error) {
	str, err := ya.execute("operation-details", bytes.NewReader([]byte("operation_id=" + operation_id)))
	if err != nil {
		return
	}
	err = json.Unmarshal(str, &ops);
	if  err != nil {
		return
	}
	return ops, err
}

func ( ya YandexMoneyClient ) OperationHistory( params ...string ) (ops []Operation, err error) {

	ready_param := ""
	for _, param := range params {
		ready_param += param
	}

	var next_record string = "0"
	for i := 0; len(next_record) > 0; i++ {

		p := ready_param

		if p != "" {
			p += "&"
		}
		p += "start_record=" + next_record

		operations, err := ya.execute("operation-history", bytes.NewReader([]byte(p)) )
		if err != nil {
			return ops, err
		}

		var res ResponseOperations
		err = json.Unmarshal(operations, &res)
		if  err != nil {
			return ops, err
		}

		for m := range res.Operations {
			ops = append( ops, res.Operations[m] )
		}

		next_record = res.NextRecord
	}

	return ops, nil
}

func NewYaMoney(token string) (YandexMoneyClient, error) {

	ya := YandexMoneyClient{
		token: token,
		api_url: "https://money.yandex.ru/api/",
	}

	return ya, nil
}

func (ya YandexMoneyClient) execute( cmd string, rbody io.Reader ) (body []byte, err error) {

	req, err := http.NewRequest("POST", ya.api_url + cmd, rbody)
	req.Header.Set("Host", "money.yandex.ru")
	req.Header.Set("Authorization", "Bearer " + ya.token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return body, err
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)

	if err!=nil{
		return body, err
	}

	return body, nil
}
