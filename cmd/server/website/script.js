async function fetchExchangeRate() {
    try {
        const response = await fetch('/api/exchange');
        if (!response.ok) {
            throw new Error('Error al obtener la tasa de cambio');
        }
        const data = await response.json();
        document.getElementById('rate').innerText = `$1 = ₡${data.rate}`;
    } catch (error) {
        document.getElementById('rate').innerText = 'Error al cargar la tasa';
        console.error(error);
    }
}

document.getElementById('refreshButton').addEventListener('click', fetchExchangeRate);

// Cargar la tasa al inicio
fetchExchangeRate();
