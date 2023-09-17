# Omg lol IP daemon

Simple daemon to monitor the IP address of a machine and update the omg.lol DNS record if it changes.

## Limitations
 - Only supports IPv4 A records
 - The record must already exist

Works for me. YMMV.

## Config

```json
{
  "email": "omg.lol@gmail.com",
  "api_key": "123123...123",
  "hostname": "whatevs",
  "username": "lolpod"
}
```