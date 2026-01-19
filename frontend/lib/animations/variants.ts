export const fadeIn = {
  hidden: { opacity: 0 },
  visible: { opacity: 1 },
}

export const slideUp = {
  hidden: { opacity: 0, y: 20 },
  visible: { opacity: 1, y: 0 },
}

export const slideDown = {
  hidden: { opacity: 0, y: -20 },
  visible: { opacity: 1, y: 0 },
}

export const slideLeft = {
  hidden: { opacity: 0, x: 20 },
  visible: { opacity: 1, x: 0 },
}

export const slideRight = {
  hidden: { opacity: 0, x: -20 },
  visible: { opacity: 1, x: 0 },
}

export const scaleIn = {
  hidden: { opacity: 0, scale: 0.95 },
  visible: { opacity: 1, scale: 1 },
}

export const scaleOut = {
  hidden: { opacity: 1, scale: 1 },
  visible: { opacity: 0, scale: 0.95 },
}

export const staggerContainer = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: {
      staggerChildren: 0.1,
    },
  },
}

export const transition = {
  duration: 0.2,
  ease: [0.4, 0, 0.2, 1],
}

export const transitionSlow = {
  duration: 0.3,
  ease: [0.4, 0, 0.2, 1],
}
