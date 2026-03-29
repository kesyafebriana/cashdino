import type { Config } from "tailwindcss";

export default {
  content: ["./app/**/*.{js,jsx,ts,tsx}", "./components/**/*.{js,jsx,ts,tsx}"],
  presets: [require("nativewind/preset")],
  theme: {
    extend: {
      colors: {
        gold: "#ffd700",
        "challenge-orange": "#e65100",
        "challenge-brown": "#bf360c",
        "gem-green": "#4cd964",
      },
    },
  },
  plugins: [],
} satisfies Config;
