# lwm2mTestServer

Bootstrap example:

bootstrap write testlwm2mclient /0/0 '[{"bn":"/0/0","n":"/0","vs":"coap://localhost:5683"},{"n":"/1","vb":true},{"n":"/2","v":3},{"n":"/10","v":111}]'
bootstrap write testlwm2mclient /0/1 '[{"bn":"/0/1","n":"/0","vs":"coap://localhost:5683"},{"n":"/1","vb":false},{"n":"/2","v":3},{"n":"/10","v":123}]'
bootstrap write testlwm2mclient /1/0 '[{"bn":"/1/0","n":"/0","v":123},{"n":"/1","v":20},{"n":"/7","vs":"U"}]'
bootstrap finish testlwm2mclient
