// snake.js
// Battlesnake logic ported from Python to pure Node.js (no dependencies)

function info() {
    return {
        color: "#754927",
        head: "alligator",
        tail: "rocket",
    };
}

function start(gameState) {
    console.log("GAME START", gameState);
}

function end(gameState) {
    console.log("GAME OVER", gameState);
}

function move(data) {
    console.log("MOVE", data);

    const you = data.you;
    const health = you.health;
    const mySize = you.length;
    const body = you.body;
    const head = [body[0].x, body[0].y];
    const walls = [data.board.width, data.board.height];
    const snakesRaw = data.board.snakes;

    const size = snakesRaw.map((s) => s.length);
    const snakes = snakesRaw.map((s) => s.body);
    const heads = snakes.map((s) => [s[0].x, s[0].y]);
    const snakeCells = snakes.flat().map((s) => [s.x, s.y]);

    let food = data.board.food.map((f) => [f.x, f.y]);
    const numFood = food.length;
    const pm = getPreviousMove(head, [body[1].x, body[1].y]);

    let moves = getRestrictions(head, mySize, walls, snakeCells, heads, size);
    let chosenMove = null;

    try {
        while (chosenMove === null) {
            // Prefer food if low health
            if (health < 45 - numFood)
                chosenMove = starving(moves, head, food);

            // Avoid walls
            if (chosenMove === null)
                chosenMove = fleeWall(moves, walls, head);

            // Attack weaker snakes
            if (chosenMove === null)
                chosenMove = killOthers(head, mySize, heads, size, moves);

            // Seek nearby food if moderately low health
            if (chosenMove === null && health < 70 - numFood)
                chosenMove = getFood(moves, head, food);

            // Continue previous direction if safe
            if (chosenMove === null && (moves.includes(pm) || moves.length === 0))
                chosenMove = pm;

            // Random fallback
            if (chosenMove === null)
                chosenMove = randomChoice(moves);

            // Prevent immediate suicide
            const nextHead = getFutureHead(head, chosenMove);
            if (
                getRestrictions(nextHead, mySize, walls, snakeCells, heads, size, false)
                    .length === 0
            ) {
                if (moves.length) {
                    moves = moves.filter((m) => m !== chosenMove);
                    if (moves.length) chosenMove = null;
                }
            }
        }
    } catch (err) {
        chosenMove = randomChoice(moves);
    }

    return { move: chosenMove, taunt: "Battle Jake!" };
}

function getFutureHead(head, move) {
    switch (move) {
        case "left":
            return [head[0] - 1, head[1]];
        case "right":
            return [head[0] + 1, head[1]];
        case "up":
            return [head[0], head[1] - 1];
        default:
            return [head[0], head[1] + 1];
    }
}

function getPreviousMove(head, second) {
    if (head[0] === second[0]) {
        return head[1] > second[1] ? "down" : "up";
    } else {
        return head[0] > second[0] ? "right" : "left";
    }
}

function fleeWall(moves, walls, head) {
    if (head[0] >= walls[0] - 2)
        return ["left", "up", "down"].find((m) => moves.includes(m));
    if (head[0] <= 1)
        return ["right", "down", "up"].find((m) => moves.includes(m));
    if (head[1] <= 1)
        return ["down", "right", "left"].find((m) => moves.includes(m));
    if (head[1] >= walls[1] - 2)
        return ["up", "left", "right"].find((m) => moves.includes(m));
}

function killOthers(head, mySize, heads, size, moves) {
    for (let i = 0; i < heads.length; i++) {
        if (size[i] < mySize) {
            const h = heads[i];
            const xdist = h[0] - head[0];
            const ydist = h[1] - head[1];

            if (Math.abs(xdist) === 1 && Math.abs(ydist) === 1) {
                if (xdist > 0 && moves.includes("right")) return "right";
                if (xdist < 0 && moves.includes("left")) return "left";
                if (ydist > 0 && moves.includes("down")) return "down";
                if (ydist < 0 && moves.includes("up")) return "up";
            } else if (
                (Math.abs(xdist) === 2 && ydist === 0) ^
                (Math.abs(ydist) === 2 && xdist === 0)
            ) {
                if (xdist === 2 && moves.includes("right")) return "right";
                if (xdist === -2 && moves.includes("left")) return "left";
                if (ydist === 2 && moves.includes("down")) return "down";
                if (moves.includes("up")) return "up";
            }
        }
    }
}

function starving(moves, head, food) {
    let move = getFood(moves, head, food);
    if (move) return move;

    for (const f of food) {
        const xdist = f[0] - head[0];
        const ydist = f[1] - head[1];
        if (
            (Math.abs(xdist) === 2 && ydist === 0) ^
            (Math.abs(ydist) === 2 && xdist === 0)
        ) {
            if (xdist === 2 && moves.includes("right")) return "right";
            if (xdist === -2 && moves.includes("left")) return "left";
            if (ydist === 2 && moves.includes("down")) return "down";
            if (ydist === -2 && moves.includes("up")) return "up";
        }
    }
}

function getFood(moves, head, food) {
    for (const f of food) {
        const xdist = f[0] - head[0];
        const ydist = f[1] - head[1];
        if (
            (Math.abs(xdist) === 1 && ydist === 0) ^
            (Math.abs(ydist) === 1 && xdist === 0)
        ) {
            if (xdist === 1 && moves.includes("right")) return "right";
            if (xdist === -1 && moves.includes("left")) return "left";
            if (ydist === 1 && moves.includes("down")) return "down";
            if (ydist === -1 && moves.includes("up")) return "up";
        }
    }
}

function getRestrictions(head, mySize, walls, snakes, heads, size, op = true) {
    const directions = { up: 1, down: 1, left: 1, right: 1 };

    // Wall avoidance
    if (head[0] === walls[0] - 1) directions.right = 0;
    else if (head[0] === 0) directions.left = 0;
    if (head[1] === 0) directions.up = 0;
    else if (head[1] === walls[1] - 1) directions.down = 0;

    // Avoid snake bodies
    for (const s of snakes) {
        const xdist = Math.abs(s[0] - head[0]);
        const ydist = Math.abs(s[1] - head[1]);
        if (xdist + ydist === 1) {
            if (xdist === 1) {
                if (s[0] > head[0]) directions.right = 0;
                else directions.left = 0;
            } else {
                if (s[1] > head[1]) directions.down = 0;
                else directions.up = 0;
            }
        }
    }

    const directions2 = { ...directions };

    // Avoid strong heads
    for (let i = 0; i < heads.length; i++) {
        if (size[i] >= mySize) {
            const h = heads[i];
            const xdist = h[0] - head[0];
            const ydist = h[1] - head[1];
            if (Math.abs(xdist) === 1 && Math.abs(ydist) === 1) {
                if (xdist > 0) directions.right = 0;
                else if (xdist < 0) directions.left = 0;
                if (ydist > 0) directions.down = 0;
                else if (ydist < 0) directions.up = 0;
            } else if (
                (Math.abs(xdist) === 2 && ydist === 0) ^
                (Math.abs(ydist) === 2 && xdist === 0)
            ) {
                if (xdist === 2) directions.right = 0;
                else if (xdist === -2) directions.left = 0;
                else if (ydist === 2) directions.down = 0;
                else directions.up = 0;
            }
        }
    }

    // Restore if forced into dead-end
    if (!Object.values(directions).includes(1) && op)
        Object.assign(directions, directions2);

    const moves = Object.keys(directions).filter((k) => directions[k] === 1);
    console.log(moves);
    return moves;
}

function randomChoice(arr) {
    if (!arr.length) return "up";
    return arr[Math.floor(Math.random() * arr.length)];
}

module.exports = { info, start, move, end };
