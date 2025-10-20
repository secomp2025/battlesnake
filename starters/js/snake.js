// Bem-vindo ao
// __________         __    __  .__                               __
// \______   \_____ _/  |__/  |_|  |   ____   ______ ____ _____  |  | __ ____
//  |    |  _/\__  \\   __\   __\  | _/ __ \ /  ___//    \\__  \ |  |/ // __ \
//  |    |   \ / __ \|  |  |  | |  |_\  ___/ \___ \|   |  \/ __ \|    <\  ___/
//  |________/(______/__|  |__| |____/\_____>______>___|__(______/__|__\\_____>
//
// Este arquivo pode ser um bom local para a lógica e funções auxiliares do seu Battlesnake.
//
// Para começar, incluímos um código para impedir que seu Battlesnake se mova para trás.
// Para mais informações, veja docs.battlesnake.com

function info() {
    return {
        color: "#F30303", // Escolha uma cor
        head: "default",  // Escolha uma cabeça
        tail: "default",  // Escolha uma cauda
    };
}

// start é chamado quando seu Battlesnake inicia uma partida
function start(gameState) {
    console.log("INÍCIO DO JOGO");
}

// end é chamado quando seu Battlesnake termina uma partida
function end(gameState) {
    console.log("FIM DO JOGO");
}

// move é chamado a cada turno e retorna o seu próximo movimento
// Movimentos válidos são "up", "down", "left" ou "right"
// Acesse https://docs.battlesnake.com/api/example-move para ver os dados disponíveis
function move(gameState) {
    const isMoveSafe = { up: true, down: true, left: true, right: true };

    // Incluímos código para impedir que sua Snake se mova para trás
    const myHead = gameState.you.body[0]; // Coordenadas da cabeça
    const myNeck = gameState.you.body[1]; // Coordenadas do "pescoço"

    if (myNeck.x < myHead.x) { // Pescoço está à esquerda da cabeça, não vá para a esquerda
        isMoveSafe.left = false;
    } else if (myNeck.x > myHead.x) { // Pescoço está à direita da cabeça, não vá para a direita
        isMoveSafe.right = false;
    } else if (myNeck.y < myHead.y) { // Pescoço está abaixo da cabeça, não vá para baixo
        isMoveSafe.down = false;
    } else if (myNeck.y > myHead.y) { // Pescoço está acima da cabeça, não vá para cima
        isMoveSafe.up = false;
    }

    // TODO: Etapa 1 - Impedir que o Battlesnake saia dos limites do tabuleiro
    // const boardWidth = gameState.board.width;
    // const boardHeight = gameState.board.height;

    // TODO: Etapa 2 - Impedir que o Battlesnake colida com ele mesmo
    // const myBody = gameState.you.body;

    // TODO: Etapa 3 - Impedir que o Battlesnake colida com outros Battlesnakes
    // const opponents = gameState.board.snakes;

    // Existem movimentos seguros disponíveis?
    const safeMoves = [];
    for (const [mv, ok] of Object.entries(isMoveSafe)) {
        if (ok) safeMoves.push(mv);
    }

    if (safeMoves.length === 0) {
        console.log(`MOVIMENTO ${gameState.turn}: Nenhum movimento seguro detectado! Indo para baixo`);
        return { move: "down" };
    }

    // Escolhe um movimento aleatório entre os seguros
    const nextMove = safeMoves[Math.floor(Math.random() * safeMoves.length)];

    // TODO: Etapa 4 - Ir em direção à comida em vez de aleatoriamente, para recuperar saúde e
    // sobreviver mais
    // const food = gameState.board.food;

    console.log(`MOVIMENTO ${gameState.turn}: ${nextMove}`);
    return { move: nextMove };
}

module.exports = { info, start, move, end };
