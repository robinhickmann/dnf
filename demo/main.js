const zone = "dns.example.com";

document.getElementById("token").textContent = generateToken().concat(".").concat(zone);

function generateToken(length = 12) {
    const chars = "0123456789abcdefghijklmnopqrstuvwxyz";
    let result = "";

    for (let i = 0; i < length; i++) {
        result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    
    return result;
}
