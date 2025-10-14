// game-client.js
// Handles live WebSocket frames, rendering, replay, and localStorage persistence.

const CELL_SIZE = 20
const CELL_SIZE_HALF = CELL_SIZE / 2
const CELL_SPACING = 4
const GRID_BORDER = 10

function makeSvgCalcParams(w, h) {
  const svgWidth = 2 * GRID_BORDER + w * CELL_SIZE + Math.max(w - 1, 0) * CELL_SPACING
  const svgHeight = 2 * GRID_BORDER + h * CELL_SIZE + Math.max(h - 1, 0) * CELL_SPACING

  return {
    cellSize: CELL_SIZE,
    cellSizeHalf: CELL_SIZE_HALF,
    cellSpacing: CELL_SPACING,
    gridBorder: GRID_BORDER,
    height: svgHeight,
    width: svgWidth,
  };
}

class Point {
  constructor(x, y) {
    this.x = x;
    this.y = y;
  }
}

class SvgCalcResult {
  constructor(x, y, width, height) {
    this.x = x;
    this.y = y;
    this.width = width;
    this.height = height;
  }
}

class SvgPoint {
  constructor(x, y) {
    this.x = x;
    this.y = y;
  }
}


function isEqualPoint(p1, p2) {
  if (p1 == undefined || p2 == undefined) {
    return false;
  }

  return p1.x == p2.x && p1.y == p2.y;
}

function isAdjacentPoint(p1, p2) {
  return calcManhattan(p1, p2) == 1;
}

function calcManhattan(p1, p2) {
  return Math.abs(p1.x - p2.x) + Math.abs(p1.y - p2.y);
}

function calcSourceWrapPosition(src, dst) {
  return {
    x: src.x - Math.sign(dst.x - src.x),
    y: src.y - Math.sign(dst.y - src.y)
  };
}

function calcDestinationWrapPosition(src, dst) {
  return {
    x: dst.x + Math.sign(dst.x - src.x),
    y: dst.y + Math.sign(dst.y - src.y)
  };
}

function svgCalcCellTopLeft(params, p) {
  const x = p.x !== undefined ? p.x : p.X;
  const y = p.y !== undefined ? p.y : p.Y;

  return new SvgPoint(
    params.gridBorder + x * (params.cellSize + params.cellSpacing),
    params.height -
    (params.gridBorder + y * (params.cellSize + params.cellSpacing) + params.cellSize)
  );
}

function svgCalcCellCenter(params, p) {
  const topLeft = svgCalcCellTopLeft(params, p);
  return {
    x: topLeft.x + params.cellSizeHalf,
    y: topLeft.y + params.cellSizeHalf
  };
}


function svgCalcCellRect(params, p) {
  const topLeft = svgCalcCellTopLeft(params, p);
  return { x: topLeft.x, y: topLeft.y, width: params.cellSize, height: params.cellSize };
}

function svgCalcCellCircle(params, p) {
  const center = svgCalcCellCenter(params, p);
  return { cx: center.x, cy: center.y };
}

const mediaCache = {};

async function fetchCustomizationSvgDef(type, name) {
  const mediaPath = `snakes/${type}s/${name}.svg`;

  if (!(mediaPath in mediaCache)) {
    mediaCache[mediaPath] = await fetch(`https://media.battlesnake.com/${mediaPath}`)
      .then((response) => response.text())
      .then((textSVG) => {
        const tempElememt = document.createElement("template");
        tempElememt.innerHTML = textSVG.trim();
        console.debug(`[customizations] loaded svg definition for ${mediaPath}`);

        if (tempElememt.content.firstChild === null) {
          console.debug("[customizations] error loading customization, no elements found");
          return "";
        }

        const child = tempElememt.content.firstChild;
        return child.innerHTML;
      });
  }
  return mediaCache[mediaPath];
}


class SvgRenderer {
  constructor(svgCanvas) {
    this.svgCanvas = svgCanvas;
    this.gameInfo = null;
  }

  setGameInfo(gameInfo) {
    // set svgViewBox
    console.log("[Render] setting game info:", gameInfo);
    this.calcParams = makeSvgCalcParams(gameInfo.Width, gameInfo.Height);
    console.log("[Render] calc params:", this.calcParams);
    this.svgCanvas.setAttribute("viewBox", `0 0 ${this.calcParams.width} ${this.calcParams.height}`);

    this.gameInfo = gameInfo;
  }

  async renderFrame(frame) {
    if (frame == null) {
      console.error("[Render] frame is undefined");
      return;
    }
    if (this.gameInfo == null) {
      console.error("[Render] game info is undefined");
      return;
    }

    if (frame.Type != "frame") {
      console.warn("[Render] frame is not a frame");
      return;
    }

    this.svgCanvas.innerHTML = "";



    this.renderGrid();

    // render eliminated snakes
    for (let snake of frame.Data.Snakes) {
      if (snake.Death != null) {
        await this.renderSnake(snake);
      }
    }

    // render snakes
    for (let snake of frame.Data.Snakes) {
      if (snake.Death == null) {
        await this.renderSnake(snake);
      }
    }

    // render food
    for (let i = 0; i < frame.Data.Food.length; i++) {
      await this.renderFood(frame.Data.Food[i], i);
    }
  }

  renderGrid() {
    let rawSvg = "<g>";
    for (let i = 0; i < this.gameInfo.Width; i++) {
      for (let j = 0; j < this.gameInfo.Height; j++) {
        const rectParams = svgCalcCellRect(this.calcParams, new Point(i, j));
        rawSvg += `<rect id="grid-${i}-${j}" class="grid fill-[#f1f1f1]"
         x="${rectParams.x}" y="${rectParams.y}" width="${rectParams.width}" height="${rectParams.height}" />`;
      }
    }
    rawSvg += "</g>";
    this.svgCanvas.innerHTML += rawSvg;
  }

  async renderFood(food, key) {
    let rawSvg = `<g id="food-${key}" class="food fill-rose-500">`;
    const circleProps = svgCalcCellCircle(this.calcParams, food);
    const foodRadius = (this.calcParams.cellSize / 3.25).toFixed(2);
    rawSvg += `<circle id="food-${key}" class="food fill-rose-500" r="${foodRadius}" ${Object.entries(circleProps).map(([key, value]) => `${key}="${value}"`).join(" ")} />`;
    rawSvg += `</g>`;
    this.svgCanvas.innerHTML += rawSvg;
  }

  async renderSnake(snake, opacity = 1.0) {
    let rawSvg = `<g id="snake-${snake.ID}" class="snake" style="opacity: ${opacity}">`;
    const tail = this.renderSnakeTail(snake);
    const body = this.renderSnakeBody(snake);
    const head = this.renderSnakeHead(snake);

    rawSvg += await tail;
    rawSvg += await body;
    rawSvg += await head;
    rawSvg += `</g>`;
    this.svgCanvas.innerHTML += rawSvg;
  }

  async renderSnakeTail(snake) {
    function calcTailTransform(snake) {
      const tail = snake.Body[snake.Body.length - 1];

      // Work backwards from the tail until we reach a segment that isn't stacked.
      let preTailIndex = snake.Body.length - 2;

      let tailPoint = new Point(tail.X, tail.Y);
      let preTailPoint = new Point(snake.Body[preTailIndex].X, snake.Body[preTailIndex].Y);


      while (preTailPoint.x == tailPoint.x && preTailPoint.y == tailPoint.y) {
        preTailIndex -= 1;
        if (preTailIndex < 0) {
          return "";
        }
        preTailPoint = new Point(snake.Body[preTailIndex].X, snake.Body[preTailIndex].Y);
      }

      // If tail is wrapped we need to calcualte neck position on border
      if (!isAdjacentPoint(preTailPoint, tailPoint)) {
        preTailPoint = calcDestinationWrapPosition(preTailPoint, tailPoint);
      }

      // Return transform based on relative location
      if (preTailPoint.x > tailPoint.x) {
        // Moving right
        return "scale(-1,1) translate(-100,0)";
      } else if (preTailPoint.y > tailPoint.y) {
        // Moving up
        return "scale(-1,1) translate(-100,0) rotate(90, 50, 50)";
      } else if (preTailPoint.y < tailPoint.y) {
        // Moving down
        return "scale(-1,1) translate(-100,0) rotate(-90, 50, 50)";
      }
      // Moving left
      return "";
    }



    const tailRectProps = svgCalcCellRect(this.calcParams, snake.Body[snake.Body.length - 1]);
    let rawSvg = `<svg class="tail" viewBox="0 0 100 100" fill="${snake.Color}" x="${tailRectProps.x}" y="${tailRectProps.y}" width="${tailRectProps.width}" height="${tailRectProps.height}">`;
    rawSvg += `<g transform="${calcTailTransform(snake)}">`;

    rawSvg += `${await fetchCustomizationSvgDef("tail", snake.TailType)}`;

    rawSvg += `</g>`;
    rawSvg += `</svg>`;
    return rawSvg;
  }

  async renderSnakeBody(snake) {
    const OVERLAP = 0.1;

    const calcBodyPolylinesPoints = (snake) => {
      // Make a copy of snake body and separate into head, tail, and body.
      const body = [...snake.Body].map(p => new Point(p.X, p.Y));
      const head = body.shift();
      const tail = body.pop();


      // Ignore body parts that are stacked on the tail
      // This ensures that the tail is always shown even when the snake has grown
      for (let last = body.at(-1); last && isEqualPoint(last, tail); last = body.at(-1)) {
        body.pop();
      }

      if (body.length == 0) {
        // If we're drawing no body, but head and tail are different,
        // they still need to be connected.
        if (!isEqualPoint(head, tail)) {
          const svgCenter = svgCalcCellCenter(this.calcParams, head);
          return [calcHeadToTailJoint(head, tail, svgCenter)];
        }

        return [[]];
      }

      return convertBodyToPolilines(body, head, tail);
    }

    const convertBodyToPolilines = (body, head, tail) => {
      const gapSize = this.calcParams.cellSpacing + OVERLAP;

      // Split wrapped body parts into separated segments
      const bodySegments = splitBodySegments(body);

      // Get the center point of each body square we're going to render
      const bodySegmentsCenterPoints = bodySegments.map((segment) =>
        segment.map(enrichSvgCellCenter)
      );

      console.log(bodySegmentsCenterPoints);

      // Extend each wrapped segment towards border
      for (let i = 0; i < bodySegmentsCenterPoints.length; i++) {
        // Extend each segment last point towards border
        if (i < bodySegmentsCenterPoints.length - 1) {
          const cur = bodySegmentsCenterPoints[i].at(-1);
          const next = bodySegmentsCenterPoints[i + 1][0];
          bodySegmentsCenterPoints[i].push(calcBorderJoint(cur, next));
        }

        // Extend segment's first point toward border portal
        if (i > 0) {
          const cur = bodySegmentsCenterPoints[i][0];
          const prev = bodySegmentsCenterPoints[i - 1].at(-1);
          bodySegmentsCenterPoints[i].unshift(calcBorderJoint(cur, prev));
        }
      }

      // Extend first point towards head
      const firstPoint = bodySegmentsCenterPoints[0][0];
      if (isAdjacentPoint(head, firstPoint)) {
        bodySegmentsCenterPoints[0].unshift(calcJoint(firstPoint, head, gapSize));
      } else {
        // Add head portal
        bodySegmentsCenterPoints[0].unshift(calcBorderJoint(enrichSvgCellCenter(firstPoint), head));
      }

      // Extend last point towards tail
      const lastPoint = bodySegmentsCenterPoints.at(-1)?.at(-1);
      if (isAdjacentPoint(lastPoint, tail)) {
        bodySegmentsCenterPoints.at(-1)?.push(calcJoint(lastPoint, tail, gapSize));
      } else {
        // Add tail portal
        bodySegmentsCenterPoints.at(-1)?.push(calcBorderJoint(lastPoint, tail));
      }

      // Finally, return an array of SvgPoints to use for a polyline
      return bodySegmentsCenterPoints.map((segment) =>
        segment.map((obj) => ({ x: obj.cx, y: obj.cy }))
      );
    }

    const splitBodySegments = (body) => {
      if (body.length == 0) {
        return [[]];
      }

      let prev = body[0];
      const segments = [[prev]];

      for (let i = 1; i < body.length; i++) {
        const cur = body[i];

        // Start new segment
        if (!isAdjacentPoint(cur, prev)) {
          segments.push([]);
        }

        segments.at(-1)?.push(cur);
        prev = cur;
      }
      return segments;
    }

    const enrichSvgCellCenter = (p) => {
      const c = svgCalcCellCenter(this.calcParams, p);
      return {
        cx: c.x,
        cy: c.y,
        ...p
      };
    }

    const calcBorderJoint = (src, dst) => {
      const border = calcSourceWrapPosition(src, dst);

      return calcJoint(src, border);
    }

    const calcJoint = (src, dst, gapSize = 0) => {
      // Extend source point towards destination
      if (dst.x > src.x) {
        return {
          ...src,
          cx: src.cx + this.calcParams.cellSizeHalf + gapSize,
          cy: src.cy
        };
      } else if (dst.x < src.x) {
        return {
          ...src,
          cx: src.cx - this.calcParams.cellSizeHalf - gapSize,
          cy: src.cy
        };
      } else if (dst.y > src.y) {
        return {
          ...src,
          cx: src.cx,
          cy: src.cy - this.calcParams.cellSizeHalf - gapSize
        };
      } else if (dst.y < src.y) {
        return {
          ...src,
          cx: src.cx,
          cy: src.cy + this.calcParams.cellSizeHalf + gapSize
        };
      }

      // In error cases there could be duplicate point
      throw new Error("Same point have no joint.");
    }

    const calcHeadToTailJoint = (head, tail, svgCenter) => {
      if (head.x > tail.x) {
        return [
          {
            x: svgCenter.x - this.calcParams.cellSizeHalf + OVERLAP,
            y: svgCenter.y
          },
          {
            x: svgCenter.x - this.calcParams.cellSizeHalf - this.calcParams.cellSpacing - OVERLAP,
            y: svgCenter.y
          }
        ];
      } else if (head.x < tail.x) {
        return [
          {
            x: svgCenter.x + this.calcParams.cellSizeHalf - OVERLAP,
            y: svgCenter.y
          },
          {
            x: svgCenter.x + this.calcParams.cellSizeHalf + this.calcParams.cellSpacing + OVERLAP,
            y: svgCenter.y
          }
        ];
      } else if (head.y > tail.y) {
        return [
          {
            x: svgCenter.x,
            y: svgCenter.y + this.calcParams.cellSizeHalf - OVERLAP
          },
          {
            x: svgCenter.x,
            y: svgCenter.y + this.calcParams.cellSizeHalf + this.calcParams.cellSpacing + OVERLAP
          }
        ];
      } else if (head.y < tail.y) {
        return [
          {
            x: svgCenter.x,
            y: svgCenter.y - this.calcParams.cellSizeHalf + OVERLAP
          },
          {
            x: svgCenter.x,
            y: svgCenter.y - this.calcParams.cellSizeHalf - this.calcParams.cellSpacing - OVERLAP
          }
        ];
      }

      throw new Error("Head and tail is a same point.");
    }


    const calcBodyPolylineProps = (polylinePoints) => {
      // Convert points into a string of the format "x1,y1 x2,y2, ...
      const points = polylinePoints
        .map((p) => {
          return `${p.x},${p.y}`;
        })
        .join(" ");

      return {
        points,
        "stroke-width": this.calcParams.cellSize,
        "stroke-linecap": "butt",
        "stroke-linejoin": "round"
      };
    }

    const bodyPolylinesPoints = calcBodyPolylinesPoints(snake);

    const drawBody = bodyPolylinesPoints[0].length > 0;
    const bodyPolylineProps = bodyPolylinesPoints.map(calcBodyPolylineProps);

    let rawSvg = "";
    if (drawBody) {
      for (let bodyPolylineProp of bodyPolylineProps) {
        rawSvg += `<polyline fill="transparent" stroke="${snake.Color}" ${Object.entries(bodyPolylineProp)
          .map(([key, value]) => `${key}="${value}"`)
          .join(" ")}/>`;
      }
    }

    return rawSvg;
  }

  async renderSnakeHead(snake) {

    function calcHeadDirection(snake) {
      const [head, neck] = snake.Body.slice(0, 2);

      let neckPoint = new Point(neck.X, neck.Y);
      let headPoint = new Point(head.X, head.Y);


      // If head is wrapped we need to calcualte neck position on border
      if (!isAdjacentPoint(neckPoint, headPoint)) {
        neckPoint.calcDestinationWrapPosition(headPoint);
      }

      // Determine head direction based on relative position of neck and tail.
      // If neck and tail overlap, we return the default direction (right).
      if (headPoint.x < neckPoint.x) {
        return "left";
      } else if (headPoint.y > neckPoint.y) {
        return "up";
      } else if (headPoint.y < neckPoint.y) {
        return "down";
      }
      return "right";
    }

    function calcHeadTransform(headDirection) {
      if (headDirection == "left") {
        return "scale(-1,1) translate(-100, 0)";
      } else if (headDirection == "up") {
        return "rotate(-90, 50, 50)";
      } else if (headDirection == "down") {
        return "rotate(90, 50, 50)";
      }

      // Moving right/default
      return "";
    }

    // If the snake is eliminated by self collision we give its head
    // a drop shadow for dramatic effect.
    function calcDrawHeadShadow(snake) {
      return snake.Death && snake.Death.Cause == "snake-self-collision";
    }

    const drawHeadShadow = calcDrawHeadShadow(snake);
    const headDirection = calcHeadDirection(snake);
    const headRectProps = svgCalcCellRect(this.calcParams, snake.Body[0]);

    let rawSvg = `<svg class="head ${headDirection} ${drawHeadShadow ? "shadow" : ""}" viewBox="0 0 100 100" fill="${snake.Color}" x="${headRectProps.x}" y="${headRectProps.y}" width="${headRectProps.width}" height="${headRectProps.height}">`;
    rawSvg += `<g transform="${calcHeadTransform(headDirection)}">`;

    rawSvg += `${await fetchCustomizationSvgDef("head", snake.HeadType)}`;

    rawSvg += `</g>`;
    rawSvg += `</svg>`;
    return rawSvg;
  }
}

const SVG_CANVAS_ID = "gameboard";

export default function initGameClient({
  framesStorageKey = "playback_frames_v1",
  gameInfoStorageKey = "playback_game_info_v1",
  autoPlay = false,
}) {
  const svgCanvas = document.getElementById(SVG_CANVAS_ID);
  if (!svgCanvas) {
    console.error("Canvas not found:", SVG_CANVAS_ID);
    return;
  }


  let ws = null;

  // ---- State ----
  let [frames, gameInfo] = loadFramesFromStorage();
  let isPlaying = autoPlay;
  let playbackIndex = 0;
  let frameTimer = null;
  let renderer = new SvgRenderer(svgCanvas);

  if (frames.length > 0 && gameInfo) {
    renderer.setGameInfo(gameInfo);

    const firstFrame = frames.find(frame => frame.Data.Turn == 1);
    if (firstFrame) {
      renderFrame(firstFrame);
    }
  }


  // ---- Live WebSocket ----

  function isConnected() {
    return ws && ws.readyState === WebSocket.OPEN;
  }

  let btnPause = document.getElementById("btn-pause");
  let btnPlay = document.getElementById("btn-play");
  let btnNextFrame = document.getElementById("btn-next-frame");
  let btnPrevFrame = document.getElementById("btn-prev-frame");
  let btnFirstFrame = document.getElementById("btn-first-frame");
  let btnLastFrame = document.getElementById("btn-last-frame");

  let onPause = () => {
    btnPause.style.display = "none";
    btnPlay.style.display = "inline";
  };

  let onPlay = () => {
    btnPause.style.display = "inline";
    btnPlay.style.display = "none";
  };


  async function connect(wsPath) {
    const wsURL =
      (location.protocol === "https:" ? "wss://" : "ws://") + location.host + wsPath;

    try {
      console.log("[WS] connecting to:", wsURL);
      ws = new WebSocket(wsURL);


      let first_frame = false;
      ws.onopen = async () => {
        console.log("[WS] connected")
        isPlaying = false;
        playbackIndex = 0;
        first_frame = true;
      };

      ws.onclose = () => {
        console.log("[WS] closed");
        saveFramesToStorage(frames);
      }
      ws.onerror = (err) => console.error("[WS] error:", err);

      ws.onmessage = async (ev) => {
        let immediateRender = false;
        if (first_frame) {
          console.log("[WS] received first frame");
          clearStorage();

          immediateRender = true;
          first_frame = false;

          try {
            const board = await fetch(`${wsPath}/events`)
            console.log("[WS] board:", board);

            gameInfo = (await board.json()).Game;
            console.log("[WS] game info:", gameInfo);

            saveGameInfoToStorage(gameInfo);

            renderer.setGameInfo(gameInfo);
          } catch (err) {
            console.error("[WS] error:", err);
          }
        }

        const frame = JSON.parse(ev.data);
        console.log("[WS] received frame:", frame);

        if (frame.Type !== "frame") {
          // game end
          return;
        }

        if (!frame.Data) {
          console.warn("[WS] received frame with no Data");
          return;
        }

        console.debug("[WS] frame turn:", frame.Data.Turn);

        // If frame already exists, replace it, otherwise insert it
        frames.push(frame);
        frames.sort((a, b) => a.Data.Turn - b.Data.Turn);

        if (immediateRender) {
          saveFramesToStorage(frames);
          console.log("[WS] rendering frame");
          await renderFrame(frame);
        }

      };
    } catch (err) {
      console.warn("[WS] error:", err);
    }
  }

  // ---- Rendering ----
  async function renderFrame(frame) {
    console.log("[Render] frame:", frame);
    await renderer.renderFrame(frame);
  }

  async function pausePlayback() {
    if (frameTimer) {
      clearInterval(frameTimer);
      frameTimer = null;
      console.log("[Playback] paused");
    }

    onPause();
  }

  async function resumePlayback(speed = 330) {
    console.log("[Playback] pausing");
    if (isPlaying || frameTimer) return;
    frameTimer = setInterval(async () => {
      console.log("[Playback] frame:", playbackIndex);
      if (playbackIndex >= frames.length) {
        if (!isConnected()) {
          console.log("[Playback] no more frames and not connected");
          stopPlayback();
          return;
        }
        // wait for next frame
        return;
      }
      await renderFrame(frames[++playbackIndex]);
    }, speed);
    console.log("[Playback] resumed");

    onPlay();
  }

  async function nextFrame() {
    if (playbackIndex >= frames.length) {
      if (!isConnected()) {
        stopPlayback();
        return;
      }
      // wait for next frame
      return;
    }
    await renderFrame(frames[++playbackIndex]);
  }

  async function prevFrame() {
    if (playbackIndex <= 0) {
      stopPlayback();
      return;
    }
    console.log("[Playback] prev frame:", frames[--playbackIndex]);
    await renderFrame(frames[playbackIndex]);
  }

  async function firstFrame() {
    playbackIndex = 0;
    await renderFrame(frames[playbackIndex]);
  }

  async function lastFrame() {
    if (playbackIndex >= frames.length) {
      if (!isConnected()) {
        stopPlayback();
        return;
      }
      // wait for next frame
      return;
    }
    playbackIndex = frames.length - 1;
    await renderFrame(frames[playbackIndex]);
  }

  function stopPlayback() {
    if (frameTimer) {
      clearInterval(frameTimer);
      frameTimer = null;
    }
    isPlaying = false;
    console.log("[Playback] stopped");

    onPause();
  }


  // ---- Storage ----
  function loadFramesFromStorage() {
    try {
      const data = localStorage.getItem(framesStorageKey);
      const info = localStorage.getItem(gameInfoStorageKey);
      if (data) {
        const arr = JSON.parse(data);
        console.log(`[Storage] Loaded ${arr.length} frames`);
        return [arr, JSON.parse(info)];
      }
    } catch (err) {
      console.warn("[Storage] load failed:", err);
    }
    return [
      [], null
    ];
  }

  function saveFramesToStorage(frames) {
    try {
      localStorage.setItem(framesStorageKey, JSON.stringify(frames));
    } catch (err) {
      console.warn("[Storage] save failed:", err);
    }
  }

  function saveGameInfoToStorage(gameInfo) {
    try {
      localStorage.setItem(gameInfoStorageKey, JSON.stringify(gameInfo));
    } catch (err) {
      console.warn("[Storage] save failed:", err);
    }
  }

  function clearStorage() {
    localStorage.removeItem(framesStorageKey);
    localStorage.removeItem(gameInfoStorageKey);
    frames = [];
    gameInfo = null;
    console.log("[Storage] cleared");
  }


  btnPause.addEventListener("click", async () => await pausePlayback());
  btnPlay.addEventListener("click", async () => await resumePlayback());
  btnNextFrame.addEventListener("click", async () => await nextFrame());
  btnPrevFrame.addEventListener("click", async () => await prevFrame());
  btnFirstFrame.addEventListener("click", async () => await firstFrame());
  btnLastFrame.addEventListener("click", async () => await lastFrame());

  if (autoPlay) {
    onPlay();
    setInterval(resumePlayback, 100);
  }

  // ---- Public API ----
  return {
    connect,
    pausePlayback,
    resumePlayback,
    clearStorage,
    get frames() {
      return frames;
    },
  };
}
