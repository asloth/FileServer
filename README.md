# FileServer

Servidor que permite transferir archivos entre 2 o más clientes usando un custom protocol

El cliente para conectarse se encuentra en el proyecto de [FileClient](https://github.com/asloth/FileClient)

## Protocol Design

Resumen del protocolo:

| Command | Argumentos | longitud | Description | 
| -- | -- | -- | -- |
| REG | @{username} | (11 bytes) | Indicamos al servidor que nos queremos registrar | 
| SUS | #{channel} | (11 bytes) | Indicamos al servidor que queremos suscribirnos a algun canal |
| UNS | #{channel} | (11 bytes) | Indicamos al servidor que queremos desuscribirnos de algun canal |
| LCH | -- | (0 bytes) | Indicamos al servidor que nos muestra la lista de canales disponibles |
| SND | #{channel} {fileSize} {fileName} | 11, 10, 64 bytes | Indicamos al servidor que queremos enviar un archivo | 

## Usos
- **Para identificarse en el servidor**

  `REG@{username}` 

  _NOTA: Username solo hasta 10 digitos._

- **Para unirse a un canal y si no existe, este se crea**

  `SUS#{channel}`

  _NOTA: Nombre del canal solo hasta 10 digitos._

- **Para ver los canales disponibles**

  `LCH`

- **Para enviar un archivo a un canal**

  `SND#{channel}{filePath}`
