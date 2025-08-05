/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ["./internal/templates/**/*.{html,js,go}", "./static/**/*.{html,js}", "./internal/components/**/*.templ"],
    theme: {
        extend: {},
    },
    plugins: [require("daisyui")],
    daisyui: {
        themes: [
            "light",
            "dark",
        ],
        darkTheme: "dark",
        base: true,
        styled: true,
        utils: true,
        prefix: "",
        logs: true,
        themeRoot: ":root",
    },
}