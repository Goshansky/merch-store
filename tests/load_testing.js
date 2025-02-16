import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = 'http://localhost:8080/api';

export let options = {
    vus: 100, // 100 виртуальных пользователей
    duration: '30s', // Тест идет 30 секунд
};

export default function () {
    let username = `testuser_${__VU}`;
    let password = 'password123';

    // 1. Регистрация пользователя
    let registerPayload = JSON.stringify({ username: username, password: password });
    let registerHeaders = { 'Content-Type': 'application/json' };
    let registerRes = http.post(`${BASE_URL}/register`, registerPayload, { headers: registerHeaders });

    check(registerRes, {
        'Регистрация успешна': (res) => res.status === 200 || res.status === 500, // 500 если уже существует
    });

    // 2. Авторизация и получение токена
    let authPayload = JSON.stringify({ username: username, password: password });
    let authRes = http.post(`${BASE_URL}/auth`, authPayload, { headers: registerHeaders });

    check(authRes, {
        'Авторизация успешна': (res) => res.status === 200,
    });

    if (authRes.status !== 200) return;

    let token = JSON.parse(authRes.body).token;
    let authHeaders = {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
    };

    // 3. Получение информации о пользователе
    let infoRes = http.get(`${BASE_URL}/info`, { headers: authHeaders });

    check(infoRes, {
        'Информация получена': (res) => res.status === 200,
    });

    let coins = JSON.parse(infoRes.body).schema.coins;

    // 4. Передача монет другому пользователю
    let sendCoinPayload = JSON.stringify({ toUser: "admin", amount: 1 });
    let sendCoinRes = http.post(`${BASE_URL}/sendCoin`, sendCoinPayload, { headers: authHeaders });

    check(sendCoinRes, {
        'Передача монет успешна': (res) => res.status === 200 || res.status === 400, // 400 если недостаточно монет
    });

    // 5. Покупка товара
    let buyPayload = JSON.stringify({ amount: 1 });
    let buyRes = http.post(`${BASE_URL}/buy/pen`, buyPayload, { headers: authHeaders });

    check(buyRes, {
        'Покупка успешна': (res) => res.status === 200,
    });


    sleep(1);
}
