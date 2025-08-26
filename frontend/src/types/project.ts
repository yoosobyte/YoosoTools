export interface Project {
    projectName?: string
    mavenSoftPath?: string
    mavenRepoPath?: string
    javaSoftPath?: string
    projectPath?: string
    autoEdu?: string
    serverUrl?: string
    serverPort?: string | number
    serverUserName?: string
    serverPassword?: string
    shellHook?: string
}

export const NewProject: Project = {
    autoEdu: '1',
    serverPort: 22,
    projectName: 'education-cloud'
}