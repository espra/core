# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""Paypal Adaptive Payments API."""

from demjson import decode as json_decode, encode as json_encode
from urllib2 import urlopen, Request as http_request, URLError

# Generic constants
PRODUCTION_ENDPOINT = 'https://svcs.paypal.com/AdaptivePayments'
SANDBOX_ENDPOINT = 'https://svcs.sandbox.paypal.com/AdaptivePayments'
API_OPERATIONS = ['Pay', 'PaymentDetails', 'Preapproval', \
    'PreapprovalDetails', 'CancelPreapproval', 'ConvertCurrency', 'Refund']


class AdaptivePayment(object):
    """PayPal Adaptive Payments API."""

    def __init__(
            self, api_username, api_password, api_signature, app_id, \
            endpoint=SANDBOX_ENDPOINT
        ):
        self._endpoint = endpoint
        self._api_username = api_username
        self._api_password = api_password
        self._api_signature = api_signature
        self._app_id = app_id

    def call(self, api_operation, api_request):
        try:
            API_OPERATIONS.index(api_operation)
        except ValueError:
            raise RuntimeError("Invalid API operation.  Please refer to"
                      "Adaptive Payments API Operations:"
                      "https://www.x.com/docs/DOC-1408#id099BG0B005Z")
        url = "%s/%s" % (self._endpoint, api_operation)
        req = http_request(url)
        # Authentication
        req.add_header('X-PAYPAL-SECURITY-USERID', self._api_username)
        req.add_header('X-PAYPAL-SECURITY-PASSWORD', self._api_password)
        req.add_header('X-PAYPAL-SECURITY-SIGNATURE', self._api_signature)
        # Specifying data formats
        req.add_header('X-PAYPAL-REQUEST-DATA-FORMAT', 'JSON')
        req.add_header('X-PAYPAL-RESPONSE-DATA-FORMAT', 'JSON')
        # Specifying application information
        req.add_header('X-PAYPAL-APPLICATION-ID', self._app_id)
        json_request = json_encode(api_request)
        try:
            openurl = urlopen(req, json_request)
        except URLError:
            raise RuntimeError("API request failed.")
        return json_decode(openurl.read())


class PayRequest(object):
    """ Creates a PayRequest message which contains the instructions required
    to make a payment.

    Only pass True for receiver_primary when making a chained payment and the
    receiver is actually the designated primary receiver for the transaction.

    You can specify the email address associated with your API credentials in
    the sender_email argument which enables PayPal to implicitly approve a
    payment without the need to redirect to PayPal.
    """

    def __init__(
            self, receiver_email, receiver_amount, currency_code, cancel_url, \
            return_url, receiver_primary=False, sender_email=False,
            preapproval_key=False, pin=False
        ):
        self._pay_request = {}
        self._pay_request['actionType'] = 'PAY'
        self._pay_request['receiverList'] = {'receiver': []}
        self._primary_set = False
        self.add_receiver(rcvr_email=receiver_email,
                rcvr_amount=receiver_amount, rcvr_primary=receiver_primary)
        self._pay_request['currencyCode'] = currency_code
        self._pay_request['cancelUrl'] = cancel_url
        self._pay_request['returnUrl'] = return_url
        self._pay_request['requestEnvelope'] = {'errorLanguage': 'en_US'}

        if sender_email:
            self._pay_request['senderEmail'] = sender_email

        if preapproval_key:
            self._pay_request['preapprovalKey'] = preapproval_key

        if pin:
            self._pay_request['pin'] = pin

    def add_receiver(self, rcvr_email, rcvr_amount, rcvr_primary):
        if len(self._pay_request['receiverList']['receiver']) < 6:
            receiver = {}
            receiver['email'] = rcvr_email
            receiver['amount'] = rcvr_amount

            if self._primary_set:
                if rcvr_primary:
                    raise RuntimeError("PayRequest already has a primary"
                                       "receiver!")
            else:
                if rcvr_primary:
                    receiver['primary'] = 'true'
                    self._primary_set = True

            self._pay_request['receiverList']['receiver'].append(receiver)
        else:
            raise RuntimeError("PayRequest has the maximum six receivers!")

    def get_message(self):
        return self._pay_request


class ResponseEnvelope(object):
    """
    Common response information, including a timestamp and the response
    acknowledgement status.
    """

    def __init__(self, pp_response):
        self._ack = pp_response['responseEnvelope']['ack']
        self._timestamp = pp_response['responseEnvelope']['timestamp']
        self._build = pp_response['responseEnvelope']['build']
        self._correlation_id = pp_response['responseEnvelope']\
                ['correlationId']


class PayResponse(object):
    """
    The PayResponse message which contains a key that you can use to identify
    the payment and the payment's status.
    """

    def __init__(self, pay_response):
        self._response_envelope = ResponseEnvelope( \
               pay_response['responseEnvelope'])
        self._error_data = ErrorData(pay_response['error'])


class PPFault(object):
    """
    The PPFault message returns ErrorData and the ResponseEnvelope
    information to your application if an error occurs when your application
    calls an Adaptive Payments API.
    """

    def __init__(self, pp_fault):
        self._response_envelope = ResponseEnvelope( \
               pp_fault['responseEnvelope'])
        self._error_data = ErrorData(pp_fault['error'])


class ErrorData(object):
    """Detailed error information returned from Paypal API."""

    def __init__(self, error_data):
        for x in error_data:
            self._error[x].category = error_data[x]['category']
            self._error[x].domain = error_data[x]['domain']
            self._error[x].error_id = error_data[x]['errorId']
            self._error[x].message = error_data[x]['message']
            self._error[x].parameter = error_data[x]['parameter']
            self._error[x].severity = error_data[x]['severity']
