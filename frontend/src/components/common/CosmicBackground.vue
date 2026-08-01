<template>
  <div ref="cosmosRef" class="cosmos" aria-hidden="true">
    <div class="cosmos-scene">
      <div class="cosmos-sky"></div>
      <div class="cosmos-color-wash"></div>
      <div class="cosmos-content-veil"></div>

      <span
        v-for="star in majorStars"
        :key="`major-${star.id}`"
        class="major-star"
        :class="`major-star--${star.color}`"
        :style="majorStarStyle(star)"
      ></span>

      <span
        v-for="star in smallStars"
        :key="`small-${star.id}`"
        class="small-star"
        :class="`small-star--${star.color}`"
        :style="smallStarStyle(star)"
      ></span>

      <div class="upper-right-composition">
        <span class="hero-flare hero-flare-main"></span>
        <span class="hero-flare hero-flare-small"></span>
        <span class="cluster-dot cluster-dot-a"></span>
        <span class="cluster-dot cluster-dot-b"></span>
        <span class="cluster-dot cluster-dot-c"></span>
        <span class="cluster-trail"></span>
      </div>

      <div class="cosmos-vignette"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

type StarColor = 'ivory' | 'cyan' | 'rose'

type MajorStar = {
  id: number
  x: number
  y: number
  size: number
  color: StarColor
  opacity: number
}

type SmallStar = {
  id: number
  x: number
  y: number
  size: number
  color: StarColor
  opacity: number
}

const cosmosRef = ref<HTMLElement | null>(null)
let frameId = 0
let reduceMotionQuery: MediaQueryList | null = null

const majorStars: MajorStar[] = [
  { id: 1, x: 59, y: 31, size: 20, color: 'ivory', opacity: 0.56 },
  { id: 2, x: 82, y: 40, size: 24, color: 'cyan', opacity: 0.66 },
  { id: 3, x: 92, y: 53, size: 22, color: 'ivory', opacity: 0.56 },
  { id: 4, x: 67, y: 60, size: 17, color: 'cyan', opacity: 0.45 },
  { id: 5, x: 14, y: 72, size: 16, color: 'rose', opacity: 0.33 },
  { id: 6, x: 53, y: 84, size: 18, color: 'cyan', opacity: 0.38 }
]

const smallStars: SmallStar[] = [
  { id: 1, x: 12, y: 24, size: 0.7, color: 'ivory', opacity: 0.26 },
  { id: 2, x: 24, y: 33, size: 0.8, color: 'cyan', opacity: 0.32 },
  { id: 3, x: 39, y: 22, size: 0.7, color: 'ivory', opacity: 0.24 },
  { id: 4, x: 50, y: 38, size: 0.8, color: 'rose', opacity: 0.27 },
  { id: 5, x: 68, y: 27, size: 0.7, color: 'ivory', opacity: 0.25 },
  { id: 6, x: 79, y: 22, size: 0.7, color: 'cyan', opacity: 0.3 },
  { id: 7, x: 95, y: 36, size: 0.8, color: 'ivory', opacity: 0.28 },
  { id: 8, x: 34, y: 48, size: 0.7, color: 'ivory', opacity: 0.22 },
  { id: 9, x: 58, y: 52, size: 0.7, color: 'cyan', opacity: 0.28 },
  { id: 10, x: 84, y: 58, size: 0.8, color: 'rose', opacity: 0.24 },
  { id: 11, x: 13, y: 64, size: 0.7, color: 'cyan', opacity: 0.24 },
  { id: 12, x: 30, y: 69, size: 0.8, color: 'ivory', opacity: 0.23 },
  { id: 13, x: 48, y: 74, size: 0.7, color: 'rose', opacity: 0.21 },
  { id: 14, x: 72, y: 78, size: 0.8, color: 'ivory', opacity: 0.24 },
  { id: 15, x: 94, y: 84, size: 0.7, color: 'cyan', opacity: 0.26 }
]

function majorStarStyle(star: MajorStar) {
  return {
    left: `${star.x}%`,
    top: `${star.y}%`,
    '--star-size': `${star.size}px`,
    '--star-opacity': star.opacity
  }
}

function smallStarStyle(star: SmallStar) {
  return {
    left: `${star.x}%`,
    top: `${star.y}%`,
    width: `${star.size}px`,
    height: `${star.size}px`,
    opacity: star.opacity
  }
}

function updateParallax() {
  frameId = 0
  if (!cosmosRef.value) return

  const reduceMotion = reduceMotionQuery?.matches ?? false
  const offset = reduceMotion ? 0 : Math.min(window.scrollY * -0.085, 0)
  cosmosRef.value.style.setProperty('--cosmos-offset', `${offset}px`)
}

function requestParallaxUpdate() {
  if (frameId) return
  frameId = window.requestAnimationFrame(updateParallax)
}

onMounted(() => {
  reduceMotionQuery = window.matchMedia('(prefers-reduced-motion: reduce)')
  updateParallax()
  window.addEventListener('scroll', requestParallaxUpdate, { passive: true })
  window.addEventListener('resize', requestParallaxUpdate, { passive: true })
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', requestParallaxUpdate)
  window.removeEventListener('resize', requestParallaxUpdate)
  if (frameId) window.cancelAnimationFrame(frameId)
})
</script>

<style scoped>
.cosmos {
  --cosmos-offset: 0px;
  position: fixed;
  inset: 0;
  z-index: 0;
  overflow: hidden;
  pointer-events: none;
  background: #010304;
}

.cosmos-scene {
  position: absolute;
  inset: -8vh -3vw;
  transform: translate3d(0, var(--cosmos-offset), 0) scale(1.045);
  transform-origin: center top;
  will-change: transform;
  backface-visibility: hidden;
}

.cosmos-sky,
.cosmos-color-wash,
.cosmos-content-veil,
.cosmos-vignette {
  position: absolute;
  inset: 0;
}

.cosmos-sky {
  background:
    linear-gradient(
      180deg,
      #010203 0%,
      #010304 25%,
      #021015 45%,
      #073039 61%,
      #554052 78%,
      #242338 92%,
      #10131c 100%
    );
}

.cosmos-color-wash {
  opacity: 0.62;
  background:
    radial-gradient(ellipse at 78% 38%, rgba(63, 205, 198, 0.08) 0%, transparent 30%),
    radial-gradient(ellipse at 17% 61%, rgba(27, 125, 126, 0.08) 0%, transparent 36%),
    radial-gradient(ellipse at 74% 78%, rgba(160, 79, 143, 0.1) 0%, transparent 38%);
}

.cosmos-content-veil {
  background:
    linear-gradient(
      90deg,
      rgba(1, 3, 5, 0.8) 0%,
      rgba(1, 3, 5, 0.62) 40%,
      rgba(1, 3, 5, 0.3) 66%,
      rgba(1, 3, 5, 0.08) 100%
    ),
    linear-gradient(
      180deg,
      rgba(1, 3, 5, 0.32) 0%,
      transparent 32%,
      transparent 70%,
      rgba(5, 5, 14, 0.2) 100%
    );
}

.major-star {
  --star-color: 240, 243, 238;
  position: absolute;
  width: 2px;
  height: 2px;
  border-radius: 50%;
  transform: translate(-50%, -50%);
  opacity: var(--star-opacity);
  background: rgb(var(--star-color));
  box-shadow:
    0 0 4px rgba(var(--star-color), 0.45),
    0 0 10px rgba(var(--star-color), 0.11);
}

.major-star::before,
.major-star::after,
.hero-flare::before,
.hero-flare::after {
  content: '';
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  border-radius: 999px;
}

.major-star::before,
.hero-flare::before {
  width: var(--star-size);
  height: 1px;
  background: linear-gradient(
    90deg,
    transparent,
    rgba(var(--star-color), 0.08) 22%,
    rgba(var(--star-color), 0.72) 47%,
    rgba(var(--star-color), 1) 50%,
    rgba(var(--star-color), 0.72) 53%,
    rgba(var(--star-color), 0.08) 78%,
    transparent
  );
}

.major-star::after,
.hero-flare::after {
  width: 1px;
  height: calc(var(--star-size) * 1.18);
  background: linear-gradient(
    180deg,
    transparent,
    rgba(var(--star-color), 0.06) 20%,
    rgba(var(--star-color), 0.7) 47%,
    rgba(var(--star-color), 1) 50%,
    rgba(var(--star-color), 0.7) 53%,
    rgba(var(--star-color), 0.06) 80%,
    transparent
  );
}

.major-star--ivory {
  --star-color: 244, 242, 232;
}

.major-star--cyan {
  --star-color: 90, 240, 233;
}

.major-star--rose {
  --star-color: 248, 162, 194;
}

.small-star {
  position: absolute;
  border-radius: 50%;
  transform: translate(-50%, -50%);
  background: currentColor;
  box-shadow: 0 0 4px currentColor;
}

.small-star--ivory {
  color: rgba(244, 242, 232, 0.9);
}

.small-star--cyan {
  color: rgba(102, 238, 232, 0.9);
}

.small-star--rose {
  color: rgba(248, 164, 196, 0.86);
}

.upper-right-composition {
  position: absolute;
  right: clamp(46px, 8vw, 126px);
  top: clamp(105px, 14vh, 176px);
  width: clamp(190px, 23vw, 310px);
  height: clamp(132px, 18vw, 220px);
  opacity: 0.92;
}

.hero-flare {
  --star-color: 91, 242, 234;
  position: absolute;
  width: 3px;
  height: 3px;
  border-radius: 50%;
  background: rgb(var(--star-color));
  box-shadow:
    0 0 7px rgba(var(--star-color), 0.54),
    0 0 24px rgba(var(--star-color), 0.14);
}

.hero-flare-main {
  --star-size: 46px;
  right: 9%;
  top: 23%;
}

.hero-flare-small {
  --star-color: 244, 242, 232;
  --star-size: 25px;
  left: 27%;
  bottom: 20%;
  opacity: 0.76;
}

.cluster-dot {
  position: absolute;
  width: 2px;
  height: 2px;
  border-radius: 50%;
  background: rgba(235, 245, 241, 0.72);
  box-shadow: 0 0 7px rgba(218, 247, 242, 0.24);
}

.cluster-dot-a {
  left: 9%;
  top: 18%;
}

.cluster-dot-b {
  left: 52%;
  top: 8%;
  background: rgba(249, 174, 204, 0.72);
}

.cluster-dot-c {
  right: 25%;
  bottom: 8%;
  background: rgba(93, 239, 233, 0.76);
}

.cluster-trail {
  position: absolute;
  left: 18%;
  top: 51%;
  width: 58%;
  height: 1px;
  opacity: 0.18;
  transform: rotate(-18deg);
  transform-origin: left center;
  background: linear-gradient(90deg, transparent, rgba(114, 230, 223, 0.52), transparent);
}

.cosmos-vignette {
  box-shadow:
    inset 0 0 190px rgba(0, 0, 0, 0.44),
    inset 0 -160px 220px rgba(4, 4, 13, 0.34);
}

@media (max-width: 720px) {
  .cosmos-scene {
    inset: -6vh -10vw;
    transform: translate3d(0, var(--cosmos-offset), 0) scale(1.08);
  }

  .upper-right-composition {
    right: 14px;
    top: 96px;
    width: 180px;
    height: 126px;
    opacity: 0.72;
  }

  .major-star:nth-of-type(n + 5),
  .small-star:nth-of-type(n + 11) {
    display: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .cosmos-scene {
    transform: scale(1.045);
  }
}
</style>
