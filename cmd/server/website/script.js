async function fetchExchangeRate() {
    try {
        const response = await fetch('/api/exchange');
        if (!response.ok) {
            throw new Error('Error al obtener la tasa de cambio');
        }
        const data = await response.json();
        const currentRate = data.rate;

        // Obtener la tasa de la cookie
        const lastRate = getCookie('last-rate');

        // Mostrar la tasa actual
        document.getElementById('rate').innerText = `₡ ${currentRate}`;

        // Comparar con la tasa anterior si existe
        if (lastRate) {
            const difference = currentRate - lastRate;
            const message = difference > 0
                ? `La tasa ha aumentado en ₡ ${difference.toFixed(2)} desde la última actualización.`
                : difference < 0
                ? `La tasa ha disminuido en ₡ ${Math.abs(difference).toFixed(2)} desde la última actualización.`
                : 'La tasa se mantiene igual.';

            showMessage(message);
        }

        // Actualizar la cookie con la nueva tasa
        setCookie('last-rate', currentRate, 1); // 1 día de expiración

    } catch (error) {
        document.getElementById('rate').innerText = 'Error al cargar la tasa';
        console.error(error);
    }
}

function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
}

function setCookie(name, value, days) {
    const expires = new Date(Date.now() + days * 864e5).toUTCString();
    document.cookie = `${name}=${value}; expires=${expires}; path=/`;
}

function showMessage(message) {
    const messageElement = document.createElement('p');
    messageElement.className = 'message';
    messageElement.innerText = message;
    document.querySelector('.container').appendChild(messageElement);
}

document.getElementById('refreshButton').addEventListener('click', fetchExchangeRate);

// Cargar la tasa al inicio
fetchExchangeRate();
