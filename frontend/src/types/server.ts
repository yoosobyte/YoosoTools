export interface Server {
    serverId?: number
    serverName?: string
    serverNickName?: string
    serverUrl?: string
    serverPort?: string | number
    serverUserName?: string
    serverPassword?: string
    sessionId?: string
}

export const NewServer: Server = {
    serverPort: '22'
}