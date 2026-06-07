import Aura from '@primeuix/themes/aura'
import { definePreset } from '@primeuix/themes'

export const loomiPreset = definePreset(Aura, {
  semantic: {
    primary: {
      50: '#eef6ff',
      500: '#2f7de1',
      700: '#1d5fb4',
    },
  },
})
