FROM node:14-alpine

WORKDIR /app

COPY package.json .
RUN npm install
RUN npm install react-toastify@latest recharts axios web-vitals
RUN npm install react-chartjs-2 chart.js
RUN npm install @babel/plugin-proposal-private-property-in-object --save-dev
RUN npm install react-datepicker



COPY . .

EXPOSE 3000

CMD ["npm", "start", "--", "--port", "3000"]