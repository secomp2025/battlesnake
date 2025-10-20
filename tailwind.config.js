/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ['./templates/pages/*.templ', './templates/components/*.templ', './static/*.js'],
    includeLanguages: { templ: "html" },
    theme: {
        extend: {},
    },
    plugins: [],
}