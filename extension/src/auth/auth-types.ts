export interface RegisterStartRequest {
    username: string;
    client_message: string;
}

export interface RegisterStartResponse {
    registration_id: string;
    server_message: string;
}

export interface RegisterFinishRequest {
    registration_id: string;
    client_message: string;
}

export interface LoginStartRequest {
    username: string;
    client_message: string;
}

export interface LoginStartResponse {
    login_id: string;
    server_message: string;
}

export interface LoginFinishRequest {
    login_id: string;
    client_message: string;
}

export interface LoginFinishResponse {
    access_token: string;
    expires_in: number;
}