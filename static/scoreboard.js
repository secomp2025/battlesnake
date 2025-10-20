function snakeIdToName(frame, id) {
    for (let i = 0; i < frame.Board.Snakes.length; i++) {
        if (frame.Board.Snakes[i].ID == id) {
            return frame.Board.Snakes[i].Name;
        }
    }
    return "DESCONHECIDO";
}

function eliminationToString(frame, elimination) {
    // Ver https://github.com/BattlesnakeOfficial/rules/blob/master/standard.go
    switch (elimination.Cause) {
        case "snake-collision":
            return `Colidiu com o corpo de ${snakeIdToName(frame, elimination.EliminatedBy)} no Turno ${elimination.Turn}`;
        case "snake-self-collision":
            return `Colidiu consigo mesma no Turno ${elimination.Turn}`;
        case "out-of-health":
            return `Perdeu todas as vidas no Turno ${elimination.Turn}`;
        case "hazard":
            return `Eliminado por obstáculo no Turno ${elimination.Turn}`;
        case "head-collision":
            return `Perdeu um jogo-de-cabeça com ${snakeIdToName(elimination.EliminatedBy)} no Turno ${elimination.Turn}`;
        case "wall-collision":
            return `Saiu dos limites no Turno ${elimination.Turn}`;
        default:
            return elimination.Cause;
    }
}

export default function initScoreboard() {
    return {
        updateScoreboard(frame) {
            console.log("[Scoreboard] updating scoreboard");
            const scoreboardTurn = document.getElementById("scoreboard-turn");
            scoreboardTurn.textContent = frame.Data.Turn;

            const sortedSnakes = frame.Data.Snakes.sort((s1, s2) => s1.Name.localeCompare(s2.Name));
            const scoreboardSnakes = document.getElementById("scoreboard-snakes");
            scoreboardSnakes.innerHTML = "";

            for (let snake of sortedSnakes) {
                console.log(snake);

                const snakeHealthHtml = snake.Death ? `<p>${eliminationToString(frame, snake.Death)}</p>` : `
                <div class="text-outline w-full h-full rounded-full bg-neutral-200 mt-2">
                    <div class="transition-all h-full rounded-full text-white ps-2" style="background: ${snake.Color}; width: ${snake.Health}%">
                        ${snake.Health}
                    </div>
                </div>`;

                let snakeHtml = `
                <div class="p-2 cursor-pointer rounded-lg border-solid border-2 border-transparent hover:border-gray-300 hover:bg-gray-200" ${snake.Death ? "class='eliminated'" : ""}
                    role="presentation">
                    <div class="flex flex-row font-bold">
                        <p class="grow truncate">${snake.Name}</p>
                        <p class="ps-4 text-right">${snake.Body.length}</p>
                    </div>
                    <div class="flex flex-row text-xs">
                        <p class="grow"></p>
                        <p class="text-right">${snake.Latency ? `${snake.Latency}ms` : ""}</p>
                    </div>
                    ${snakeHealthHtml}
                </div>`;

                scoreboardSnakes.innerHTML += snakeHtml;
            }
        }
    }
}