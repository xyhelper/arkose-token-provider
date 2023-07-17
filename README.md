# arkose-token-provider

```yaml
services:
  arkose-token-provider:
    container_name: arkose-token-provider
    image: linweiyuan/arkose-token-provider
    ports:
      - 8080:8080
    environment:
      - TZ=Asia/Shanghai
      - BX=
      - INTERVAL=3
    restart: unless-stopped
```
