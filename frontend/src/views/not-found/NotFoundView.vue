<template>
  <main class="not-found">
    <section class="not-found__content">
      <div class="not-found__header">
        <h1>404</h1>
        <p>页面找不到了，又有一个程序员要失业了。</p>
      </div>

      <div class="dino-game">
        <div class="dino-game__top">
          <span>{{ statusText }}</span>
          <strong>Score {{ score }}</strong>
        </div>

        <canvas
          ref="canvasRef"
          class="dino-game__canvas"
          width="720"
          height="220"
          @click="handleCanvasClick"
        />

        <div class="dino-game__footer">
          <span>空格 / ↑ 跳跃，Enter 重新开始</span>
          <span>Best {{ bestScore }}</span>
        </div>
      </div>

      <el-button type="primary" @click="$router.push('/')">返回首页</el-button>
    </section>
  </main>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

const GAME_WIDTH = 720
const GAME_HEIGHT = 220
const GROUND_Y = 166
const GRAVITY = 1800
const JUMP_SPEED = -640
const PIXEL = 4
const DINO_COLOR = '#535353'
const GROUND_COLOR = '#6f6f6f'

const canvasRef = ref()
const score = ref(0)
const bestScore = ref(0)
const gameStatus = ref('ready')

const dino = {
  x: 58,
  y: GROUND_Y - 48,
  width: 58,
  height: 48,
  velocityY: 0,
  onGround: true,
}

let speed = 320
let lastTime = 0
let distance = 0
let nextObstacleDistance = 460
let obstacles = []
let animationId = 0

const statusText = computed(() => {
  if (gameStatus.value === 'gameover') return '撞到了，程序员正在写离职申请'
  if (gameStatus.value === 'running') return '正在寻找丢失的页面'
  return '按空格开始'
})

const getCanvasContext = () => {
  return canvasRef.value?.getContext('2d')
}

const drawPixelRect = (ctx, x, y, width, height, scale = PIXEL) => {
  ctx.fillRect(Math.round(x), Math.round(y), width * scale, height * scale)
}

const drawPixelRects = (ctx, x, y, rects, scale = PIXEL) => {
  rects.forEach(([rectX, rectY, width, height]) => {
    drawPixelRect(ctx, x + rectX * scale, y + rectY * scale, width, height, scale)
  })
}

const drawDino = (ctx) => {
  const runFrame = Math.floor(distance / 48) % 2
  const legRects =
    dino.onGround && gameStatus.value === 'running'
      ? runFrame === 0
        ? [
            [4, 11, 2, 3],
            [4, 14, 1, 1],
            [8, 11, 2, 2],
            [9, 13, 1, 2],
            [9, 15, 2, 1],
          ]
        : [
            [4, 11, 2, 2],
            [5, 13, 1, 2],
            [4, 15, 2, 1],
            [8, 11, 2, 3],
            [9, 14, 1, 1],
          ]
      : [
          [4, 11, 2, 4],
          [8, 11, 2, 4],
        ]

  ctx.fillStyle = DINO_COLOR
  drawPixelRects(ctx, dino.x, dino.y, [
    // Body and back 身体和背部
    [3, 7, 7, 4],
    [4, 6, 5, 1],
    [2, 8, 1, 2],
    [9, 8, 2, 2],
    [10, 9, 1, 2],
    // Neck and head 脖子和头部
    [8, 4, 2, 3],
    [9, 1, 5, 4],
    [10, 0, 4, 1],
    [14, 2, 1, 2],
    [13, 4, 2, 1],
    // Tail and arm 尾巴和短手
    [0, 6, 3, 1],
    [1, 7, 2, 1],
    [9, 7, 3, 1],
    [11, 8, 1, 1],
    ...legRects,
  ])

  ctx.fillStyle = '#ffffff'
  drawPixelRect(ctx, dino.x + 12 * PIXEL, dino.y + 1 * PIXEL, 1, 1)
}

const drawObstacle = (ctx, obstacle) => {
  ctx.fillStyle = DINO_COLOR
  obstacle.parts.forEach((part) => {
    drawPixelRects(ctx, obstacle.x + part.x, obstacle.y + part.y, part.rects)
  })
}

const drawScene = () => {
  const ctx = getCanvasContext()
  if (!ctx) return

  ctx.clearRect(0, 0, GAME_WIDTH, GAME_HEIGHT)

  ctx.fillStyle = '#ffffff'
  ctx.fillRect(0, 0, GAME_WIDTH, GAME_HEIGHT)

  ctx.fillStyle = '#f1f1f1'
  ctx.fillRect(0, 0, GAME_WIDTH, GAME_HEIGHT)

  ctx.fillStyle = '#ffffff'
  ctx.fillRect(0, 0, GAME_WIDTH, GROUND_Y + 3)

  ctx.fillStyle = GROUND_COLOR
  for (let x = -((distance / 3) % 46); x < GAME_WIDTH; x += 46) {
    ctx.fillRect(x, GROUND_Y + 8, 18, 2)
    ctx.fillRect(x + 28, GROUND_Y + 14, 8, 2)
  }

  ctx.strokeStyle = GROUND_COLOR
  ctx.lineWidth = 1
  ctx.beginPath()
  ctx.moveTo(0, GROUND_Y)
  ctx.lineTo(GAME_WIDTH, GROUND_Y)
  ctx.stroke()

  drawClouds(ctx)
  obstacles.forEach((obstacle) => drawObstacle(ctx, obstacle))
  drawDino(ctx)
}

const drawClouds = (ctx) => {
  ctx.strokeStyle = '#c7c7c7'
  ctx.lineWidth = 2
  const cloudOffset = (distance / 8) % GAME_WIDTH
  ;[170, 470, 760].forEach((baseX, index) => {
    const x = baseX - cloudOffset
    const cloudX = x < -90 ? x + GAME_WIDTH + 180 : x
    const y = 42 + index * 18
    ctx.beginPath()
    ctx.moveTo(cloudX, y + 16)
    ctx.lineTo(cloudX + 18, y + 16)
    ctx.lineTo(cloudX + 24, y + 8)
    ctx.lineTo(cloudX + 38, y + 8)
    ctx.lineTo(cloudX + 46, y)
    ctx.lineTo(cloudX + 62, y)
    ctx.lineTo(cloudX + 70, y + 8)
    ctx.lineTo(cloudX + 88, y + 8)
    ctx.lineTo(cloudX + 96, y + 16)
    ctx.lineTo(cloudX + 112, y + 16)
    ctx.stroke()
  })
}

const createCactusParts = (variant) => {
  const cactus = [
    {
      x: 0,
      y: 0,
      rects: [
        [2, 0, 2, 11],
        [1, 1, 4, 1],
        [0, 5, 2, 2],
        [0, 4, 1, 1],
        [4, 3, 2, 2],
        [5, 2, 1, 1],
      ],
    },
  ]
  if (variant > 0) {
    cactus.push({
      x: 18,
      y: 12,
      rects: [
        [2, 0, 2, 8],
        [1, 1, 4, 1],
        [4, 3, 2, 2],
        [5, 2, 1, 1],
      ],
    })
  }
  if (variant > 1) {
    cactus.push({
      x: 38,
      y: 6,
      rects: [
        [2, 0, 2, 9],
        [1, 1, 4, 1],
        [0, 4, 2, 2],
        [0, 3, 1, 1],
      ],
    })
  }
  return cactus
}

const createObstacle = () => {
  const variant = Math.floor(Math.random() * 3)
  const width = 24 + variant * 20
  const height = 44

  obstacles.push({
    x: GAME_WIDTH + 20,
    y: GROUND_Y - height,
    width,
    height,
    parts: createCactusParts(variant),
  })
}

const hasCollision = (obstacle) => {
  const dinoBox = {
    x: dino.x + 8,
    y: dino.y + 4,
    width: dino.width - 14,
    height: dino.height - 6,
  }

  return (
    dinoBox.x < obstacle.x + obstacle.width &&
    dinoBox.x + dinoBox.width > obstacle.x &&
    dinoBox.y < obstacle.y + obstacle.height &&
    dinoBox.y + dinoBox.height > obstacle.y
  )
}

const stopGame = () => {
  gameStatus.value = 'gameover'
  bestScore.value = Math.max(bestScore.value, score.value)
  cancelAnimationFrame(animationId)
}

const updateGame = (time) => {
  if (gameStatus.value !== 'running') return

  const deltaTime = Math.min((time - lastTime) / 1000, 0.032)
  lastTime = time
  distance += speed * deltaTime
  speed += 12 * deltaTime
  score.value = Math.floor(distance / 12)

  dino.velocityY += GRAVITY * deltaTime
  dino.y += dino.velocityY * deltaTime

  if (dino.y >= GROUND_Y - dino.height) {
    dino.y = GROUND_Y - dino.height
    dino.velocityY = 0
    dino.onGround = true
  }

  nextObstacleDistance -= speed * deltaTime
  if (nextObstacleDistance <= 0) {
    createObstacle()
    nextObstacleDistance = 300 + Math.random() * 280
  }

  obstacles = obstacles
    .map((obstacle) => ({
      ...obstacle,
      x: obstacle.x - speed * deltaTime,
    }))
    .filter((obstacle) => obstacle.x + obstacle.width > -20)

  if (obstacles.some(hasCollision)) {
    stopGame()
    drawScene()
    return
  }

  drawScene()
  animationId = requestAnimationFrame(updateGame)
}

const jump = () => {
  if (!dino.onGround) return

  dino.velocityY = JUMP_SPEED
  dino.onGround = false
}

const startGame = () => {
  cancelAnimationFrame(animationId)
  gameStatus.value = 'running'
  score.value = 0
  speed = 320
  distance = 0
  nextObstacleDistance = 460
  obstacles = []
  dino.y = GROUND_Y - dino.height
  dino.velocityY = 0
  dino.onGround = true
  lastTime = performance.now()
  drawScene()
  animationId = requestAnimationFrame(updateGame)
}

const handleAction = () => {
  if (gameStatus.value === 'running') {
    jump()
    return
  }

  startGame()
}

const handleKeydown = (event) => {
  if (event.code === 'Space' || event.code === 'ArrowUp') {
    event.preventDefault()
    handleAction()
  }

  if (event.code === 'Enter') {
    startGame()
  }
}

const handleCanvasClick = () => {
  handleAction()
}

onMounted(() => {
  drawScene()
  window.addEventListener('keydown', handleKeydown)
})

onBeforeUnmount(() => {
  cancelAnimationFrame(animationId)
  window.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped lang="scss">
.not-found {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100%;
  padding: 24px;
  color: #303133;
}

.not-found__content {
  display: flex;
  align-items: center;
  flex-direction: column;
  width: min(820px, 100%);
  padding: 0;
  text-align: center;
  transform: translateY(-24px);
}

.not-found__header {
  margin-bottom: 24px;

  h1 {
    margin: 0 0 8px;
    font-size: 64px;
    line-height: 1;
  }

  p {
    margin: 0;
    color: #909399;
    font-size: 16px;
  }
}

.dino-game {
  width: 100%;
  margin-bottom: 24px;
  overflow: hidden;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
}

.dino-game__top,
.dino-game__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 16px 20px;
  color: #909399;
  font-size: 13px;
}

.dino-game__top {
  border-bottom: 1px solid #e4e7ed;
}

.dino-game__footer {
  border-top: 1px solid #e4e7ed;
}

.dino-game__canvas {
  display: block;
  width: 100%;
  height: auto;
  cursor: pointer;
}

@media (max-width: 640px) {
  .not-found__content {
    padding: 16px;
  }

  .dino-game__top,
  .dino-game__footer {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
