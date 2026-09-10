CreateApp
curl -k -X POST "https://merchantapi.easypay.ua/api/system/createApp"   -H "Content-Type: application/json"   -H "PartnerKey: easypay-test"   -H "locale: ua"   -d "{}"

{"logoPath":"https://cdn.easypay.ua/logo/","hintImagesPath":"https://cdn.easypay.ua/hint_images/","apiVersion":"1.0","appId":"9414a644-1abe-4579-ab80-fa3d1e928548","pageId":"4350cfbf-0755-44ab-8f7d-980a1ac8d015","requestedSessionId":"4350cfbf-0755-44ab-8f7d-980a1ac8d015","error":null}

CreatePage
curl -k -X POST "https://merchantapi.easypay.ua/api/system/createPage"   -H "Content-Type: application/json"   -H "PartnerKey: easypay-test"  -H "AppId: 9414a644-1abe-4579-ab80-fa3d1e928548"   -H "locale: ua"   -d "{}"

CreateOrder
curl -k -X POST "https://merchantapi.easypay.ua/api/merchant/createOrder"   -H "Content-Type: application/json" -H "PartnerKey: easypay-test" -H "AppId:  9414a644-1abe-4579-ab80-fa3d1e928548"  -H "Sign: 7LKwnESc7rUh6zM/Lu/bjKWEbD2uo0GTP54mtbSKBjU=" -H "PageId: 4e3566d9-6068-4d29-880e-685bbb76ffd7"  -H "locale: ua"   -d '{"order":{"serviceKey":"MERCHANT-TEST","orderId":"test_20260908-0045","description":"Test payment","amount":"1","additionalItems":{"Merchant.UrlNotify":"https://tidy-worsening-womanless.ngrok-free.dev/notify"}}}'


ConfirmCode
curl -k -X POST "https://merchantapi.easypay.ua/api/payment/confirmCodeVerification"   -H "Content-Type: application/json" -H "PartnerKey: easypay-test" -H "AppId:  9414a644-1abe-4579-ab80-fa3d1e928548"  -H "Sign: lDyBNOy7ZY/Pi1Fboy6+jmS6mndyo4ccuWHSMeTXLeo=" -H "PageId: 4e3566d9-6068-4d29-880e-685bbb76ffd7"  -H "locale: ua"   -d '{ "code": "123123"}'


2026/09/08 14:16:52 ============================
2026/09/08 14:27:22 ========== NOTIFY ==========
2026/09/08 14:27:22 Method: POST
2026/09/08 14:27:22 URI: http://tidy-worsening-womanless.ngrok-free.dev/notify
2026/09/08 14:27:22 RemoteAddr: 127.0.0.1:50876
2026/09/08 14:27:22 Headers:
2026/09/08 14:27:22 Host: tidy-worsening-womanless.ngrok-free.dev
2026/09/08 14:27:22 Content-Length: 744
2026/09/08 14:27:22 Content-Type: application/json; charset=utf-8
2026/09/08 14:27:22 Accept: application/json
2026/09/08 14:27:22 Sign: jWtewuYA0xvpMXTUWN0tfCVCYVYNBO/naThq0i5+msA=
2026/09/08 14:27:22 X-Forwarded-For: 93.183.196.26
2026/09/08 14:27:22 X-Forwarded-Host: tidy-worsening-womanless.ngrok-free.dev
2026/09/08 14:27:22 X-Forwarded-Proto: https
2026/09/08 14:27:22 Accept-Encoding: gzip
2026/09/08 14:27:22 Body:
2026/09/08 14:27:22 {"OperationType":"Payment","PartnerKey":"easypay-test","ServiceKey":"MERCHANT-TEST","TransactionStatus":"Declined","MerchantOrderId":"test_20260908-002","DateTime":"2026-09-08T14:24:18","Amount":1.00,"Commission":0.00,"TransactionId":1859083939,"AdditionalItems":{"Account":"test_20260908-002","Acquirer.Name":"Raiffeisen Bank Aval ","Acquirer.TerminalId":"E0100604","Card.BrandType":"MasterCard","Card.Pan":"51678032****8169","Description":"Test payment","ErrorCode":"PAYMENT_UPC_AUTH_FAILED","ErrorMessage":"Помилка підтвердження власника картки (3D secure)","Merchant.IsOneTimePay":"True","Merchant.OrderId":"test_20260908-002","Merchant.UrlNotify":"https://tidy-worsening-womanless.ngrok-free.dev/notify"}}
2026/09/08 14:27:22 ============================

/api/system/createApp
/api/system/createPage
/api/merchant/createOrder
/api/merchant/tokenCard/create






curl -X POST https://api.rozetkapay.com/api/payments/v1/new   -u "a6a29002-dc68-4918-bc5d-51a6094b14a8:XChz3J8qrr"   -H "Content-Type: application/json"   -d '{
    "external_id": "order_1",
    "amount": 100,
    "currency": "UAH",
    "mode": "hosted",
    "callback_url": "https://tidy-worsening-womanless.ngrok-free.dev/notify",
    "result_url": "https://your-site.com/result"
  }'

{"id":"407932919957434368","external_id":"order_1","has_child":false,"project_id":"f4e62d2b-39f1-4c98-82cc-9aebdd88b579","is_success":true,"details":null,"receipt_url":null,"action_required":true,"action":{"type":"url","value":"https://buy.rozetkapay.com/order/2076e5d7-9907-48e8-b08f-849e3bfedf03?color_mode=light"},"payment_method":null,"customer":null,"operation":"payment","metadata":null}

curl -X POST "https://api.rozetkapay.com/api/customers/v1/wallet?external_id=user_123"   -u "a6a29002-dc68-4918-bc5d-51a6094b14a8:XChz3J8qrr"   -H "Content-Type: application/json"   -d '{
    "mode": "hosted",
    "callback_url": "https://tidy-worsening-womanless.ngrok-free.dev/notify",
    "result_url": "https://your-site.com/result",
    "make_default": true,
    "payment_method": {
      "type": "cc_token",
      "cc_token": {
        "token": "tok_card_token",
        "mask": "424242******4242",
        "expires_at": "2027-12-01T00:00:00Z",
        "use_3ds_flow": true
      }
    }
  }'


  curl -X POST "https://api.rozetkapay.com/api/customers/v1/wallet" \
  -u "a6a29002-dc68-4918-bc5d-51a6094b14a8:XChz3J8qrr" \
  -H "X-CUSTOMER-AUTH: user_1234" \
  -H "Content-Type: application/json" \
  -d '{
    "mode": "direct",
    "callback_url": "https://tidy-worsening-womanless.ngrok-free.dev/notify",
    "result_url": "https://your-site.com/result",
    "make_default": true,
    "payment_method": {
      "type": "cc_token",
      "cc_token": {
        "token": "NWQyN2Y3ZmZiZjU1NDM5ZWIzYTFjYWVjY2VhMzJlZTc6bmtDb21LYjFmM3p1Yk1nMFk4",
        "mask": "42424242****4242",
        "expires_at": "2031-04-30T00:00:00Z",
        "use_3ds_flow": true
      }
    }
  }'

  curl -X POST "https://api.rozetkapay.com/api/customers/v1/wallet?external_id=user_1234" \
  -u "a6a29002-dc68-4918-bc5d-51a6094b14a8:XChz3J8qrr" \
  -H "Content-Type: application/json" \
  -d '{
    "mode": "hosted",
    "callback_url": "https://tidy-worsening-womanless.ngrok-free.dev/notify",
    "result_url": "https://your-site.com/result",
    "make_default": true
  }'

  "payment_method":{"type":"cc_token","cc_token":{"token":"NWQyN2Y3ZmZiZjU1NDM5ZWIzYTFjYWVjY2VhMzJlZTc6bmtDb21LYjFmM3p1Yk1nMFk4","mask":"42424242****4242","expires_at":"2031-04-30T00:00:00Z","bank_short_name":null,"payment_system":"VISA","saved_card":false,"bin_country":"United Kingdom of Great Britain and Northern Ireland"}}

-- добавление картьі в систему доступно ТОЛЬКО при нажатии пользователя на платежной системе "сохранить карту" и тогда в нотифай по идее придет єтот cc_card_token


!!! НЕРАБОЧИЙ ВАРИАНТ ДЛЯ ДОБАВЛЕНИЕ ТЕСТОВОЙ КАРТЬІ - INTERNAL ERROR
curl -X POST "https://api.rozetkapay.com/api/customers/v1/wallet?external_id=user_123" \
  -u "a6a29002-dc68-4918-bc5d-51a6094b14a8:XChz3J8qrr" \
  -H "Content-Type: application/json" \
  -d '{
    "mode": "hosted",
    "callback_url": "https://tidy-worsening-womanless.ngrok-free.dev/notify",
    "result_url": "https://your-site.com/result",
    "make_default": true
  }'
{"code":"internal_error","type":"customer_error","message":"Internal error occurred. Please, contact technical support.","payment_id":"","param":""}

TODO: поднять колбек сервис и сделать тест пеймент с нажатием кнопки "сохранить картку" и проверить запрос от розетки в колбек сервисе
TODO: RozetkaPayClient - /payments/new (вероятно с поддержкой нажатия кнопки "сохранить карту" юзеров и вьітаскиванием токена после єтого в колбек сервисе)

curl https://api.rozetkapay.com/api/customers/v1/wallet?external_id=user_123 \
  -X POST \
  -H 'Content-Type: application/json' \
  -u "a6a29002-dc68-4918-bc5d-51a6094b14a8:XChz3J8qrr" \
  -d '{
  "callback_url": "https://tidy-worsening-womanless.ngrok-free.dev/notify",
  "result_url": "https://your-site.com/result",
  "make_default": false,
  "mode": "hosted"
}'


curl -X GET "https://api.rozetkapay.com/api/customers/v1/wallet?external_id=user_123" \
  -u "a6a29002-dc68-4918-bc5d-51a6094b14a8:XChz3J8qrr" \


curl -X POST https://api.rozetkapay.com/api/payments/v1/new   -u "a6a29002-dc68-4918-bc5d-51a6094b14a8:XChz3J8qrr"   -H "Content-Type: application/json"   -d '{
    "external_id": "order_13412312",
    "amount": 1,
    "currency": "UAH",
    "mode": "hosted",
    "callback_url": "https://tidy-worsening-womanless.ngrok-free.dev/notify",
    "result_url": "https://your-site.com/result"
  }'


curl -X POST https://api.rozetkapay.com/api/payments/v1/new   -u "a6a29002-dc68-4918-bc5d-51a6094b14a8:XChz3J8qrr"   -H "Content-Type: application/json"   -d '{
    "external_id": "order_13412312",
    "amount": 1,
    "currency": "UAH",
    "mode": "direct",
    "callback_url": "https://tidy-worsening-womanless.ngrok-free.dev/notify",
    "result_url": "https://your-site.com/result",
    "customer":{
      "external_id": "user_13412312",
      "payment_method":{
        "type":"cc_token",
        "cc_token":{
          "token":"MWMzZmZjN2RhZDRiNDY3YWI2NzM0MzE4N2JjYzJmNGE6Nld6RlBOcWd1cnZ0VFFNcVZo",
          "mask":"42424242****4242",
          "expires_at":"2029-05-31T00:00:00Z",
          "bank_short_name":null,
          "payment_system":"VISA",
          "saved_card":false,
          "bin_country":"United Kingdom of Great Britain and Northern Ireland"
        }
      }
    }
  }'

"payment_method":{"type":"cc_token","cc_token":{"token":"MWMzZmZjN2RhZDRiNDY3YWI2NzM0MzE4N2JjYzJmNGE6Nld6RlBOcWd1cnZ0VFFNcVZo","mask":"42424242****4242","expires_at":"2029-05-31T00:00:00Z","bank_short_name":null,"payment_system":"VISA","saved_card":false,"bin_country":"United Kingdom of Great Britain and Northern Ireland"}}

"payment_method":{"type":"cc_token","cc_token":{"token":"MWMzZmZjN2RhZDRiNDY3YWI2NzM0MzE4N2JjYzJmNGE6Nld6RlBOcWd1cnZ0VFFNcVZo","mask":"42424242****4242","expires_at":"2029-05-31T00:00:00Z","bank_short_name":null,"payment_system":"VISA","saved_card":false,"bin_country":"United Kingdom of Great Britain and Northern Ireland"}}

curl https://api.rozetkapay.com/api/customers/v1/wallet?external_id=user_11235312413 \
  -X POST \
  -u "a6a29002-dc68-4918-bc5d-51a6094b14a8:XChz3J8qrr" \
  -H 'Content-Type: application/json' \
  -d '{
  "mode": "direct",
  "make_default": true,
  "callback_url": "https://tidy-worsening-womanless.ngrok-free.dev/notify",
  "result_url": "https://your-site.com/result",
    "payment_method":{ 
      "type":"cc_token",
      "cc_token":{
        "token":"MWMzZmZjN2RhZDRiNDY3YWI2NzM0MzE4N2JjYzJmNGE6Nld6RlBOcWd1cnZ0VFFNcVZo",
        "mask":"42424242****4242",
        "expires_at":"2029-05-31T00:00:00Z",
        "bank_short_name":null,
        "payment_system":"VISA",
        "saved_card":true,
        "bin_country":"United Kingdom of Great Britain and Northern Ireland"
    }
  }
}'

curl https://api.rozetkapay.com/api/customers/v1/wallet \
  -X POST \
  -u "a6a29002-dc68-4918-bc5d-51a6094b14a8:XChz3J8qrr" \
  -H "X-CUSTOMER-AUTH: user_11235312413" \
  -H 'Content-Type: application/json' \
  -d '{
  "mode": "direct",
  "make_default": true,
  "callback_url": "https://tidy-worsening-womanless.ngrok-free.dev/notify",
  "result_url": "https://your-site.com/result",
  "payment_method":{ 
    "type":"cc_token",
    "cc_token":{
      "token":"MWMzZmZjN2RhZDRiNDY3YWI2NzM0MzE4N2JjYzJmNGE6Nld6RlBOcWd1cnZ0VFFNcVZo",
      "mask":"42424242****4242",
      "expires_at":"2029-05-31T00:00:00Z",
      "bank_short_name":null,
      "payment_system":"VISA",
      "saved_card":true,
      "bin_country":"United Kingdom of Great Britain and Northern Ireland"
    }
  }
}'

curl 'https://api.rozetkapay.com/api/customers/v1/wallet' \
  -X POST \
  -u "a6a29002-dc68-4918-bc5d-51a6094b14a8:XChz3J8qrr" \
  -H 'X-CUSTOMER-AUTH: user_11235312413' \
  -H 'Content-Type: application/json' \
  -d '{
    "mode": "direct",
    "payment_method": {
      "type": "cc_token",
      "cc_token":{
        "token":"MWMzZmZjN2RhZDRiNDY3YWI2NzM0MzE4N2JjYzJmNGE6Nld6RlBOcWd1cnZ0VFFNcVZo",
        "mask":"42424242****4242",
        "expires_at":"2029-05-31T00:00:00Z",
        "bank_short_name":"PrivatBank",
        "payment_system":"VISA",
        "bin_country":"United Kingdom of Great Britain and Northern Ireland"
      }
    }
  }'
