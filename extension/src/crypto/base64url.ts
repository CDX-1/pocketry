export function bytesToBase64URL(bytes: Uint8Array): string {
    let binary = "";

    for (const byte of bytes) {
        binary += String.fromCharCode(byte);
    }

    return btoa(binary)
        .replace(/\+/g, "-")
        .replace(/\//g, "_")
        .replace(/=+$/g, "")
}

export function base64URLToBytes(value: string): Uint8Array {
    const normalized = value.replace(/-/g, "+").replace(/_/g, "/");
    const paddingLength = (4 - (normalized.length % 4)) % 4;
    const padded = normalized + "=".repeat(paddingLength);

    const binary = atob(padded);
    const result = new Uint8Array(binary.length);

    for (let index = 0; index < binary.length; index++) {
        result[index] = binary.charCodeAt(index);
    }

    return result;
}