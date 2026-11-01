const API_URL = '/players';

async function fetchPlayers() {
    try {
        const response = await fetch(API_URL);
        const players = await response.json();
        renderPlayers(players);
    } catch (error) {
        console.error("Ошибка при загрузке игроков:", error);
    }
}

function renderPlayers(players) {
    const container = document.getElementById('playersContainer');
    container.innerHTML = '';

    if (!players || players.length === 0) {
        container.innerHTML = '<p style="color: #818C99;">В лобби пока никого нет.</p>';
        return;
    }

    players.forEach(player => {
        const card = document.createElement('div');
        card.className = 'player-card';
        card.innerHTML = `
            <div class="player-info">
                <span class="player-name">${player.name}</span>
                <span class="player-level">Уровень ${player.level}</span>
            </div>
            <button class="delete-btn" onclick="deletePlayer('${player.name}')" title="Удалить игрока">
                <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                    <path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/>
                </svg>
            </button>
        `;
        container.appendChild(card);
    });
}

// Функция удаления
async function deletePlayer(name) {
    if (!confirm(`Удалить игрока ${name}?`)) return;

    try {
        const response = await fetch(`${API_URL}?name=${encodeURIComponent(name)}`, {
            method: 'DELETE'
        });

        if (response.ok) {
            fetchPlayers(); // Обновляем список
        } else {
            alert("Не удалось удалить игрока");
        }
    } catch (error) {
        console.error("Ошибка при удалении:", error);
    }
}

document.getElementById('playerForm').addEventListener('submit', async (e) => {
    e.preventDefault();

    const name = document.getElementById('playerName').value;
    const level = parseInt(document.getElementById('playerLevel').value);

    try {
        const response = await fetch(API_URL, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, level })
        });

        if (response.ok) {
            document.getElementById('playerName').value = '';
            document.getElementById('playerLevel').value = '1';
            fetchPlayers();
        }
    } catch (error) {
        console.error("Ошибка при добавлении:", error);
    }
});

fetchPlayers();