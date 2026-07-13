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

export interface RegisterFinishResponse {
    status: string;
    user_id: number;
    username: string;
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

export interface LoginResult extends LoginFinishResponse {
    server_static_public_key: string;
}

export interface RegistrationResult {
    server_static_public_key: string;
}

export interface CurrentUserResponse {
    id: number;
    username: string;
}