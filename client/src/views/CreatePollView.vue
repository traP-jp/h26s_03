<template>
  <div class="page-container">
    <div class="header">
      <router-link to="/" class="back-link">＜ 戻る</router-link>
      <h1>新しい投票を作成</h1>
    </div>
    <div class="form-container">
      <div class="form-group">
        <label for="poll-name" class="form-label">投票名</label>
        <input
          id="poll-name"
          v-model="form.name"
          type="text"
          class="form-input"
          placeholder="例) W杯 日本 vs チュニジア"
          required
        />
      </div>
      <div class="form-group">
        <label for="choice1" class="form-label">選択肢1</label>
        <input
          id="choice1"
          v-model="form.choice1"
          type="text"
          class="form-input"
          placeholder="選択肢1"
          required
        />
      </div>
      <div class="form-group">
        <label for="choice2" class="form-label">選択肢2</label>
        <input
          id="choice2"
          v-model="form.choice2"
          type="text"
          class="form-input"
          placeholder="選択肢2"
          required
        />
      </div>
      <div class="form-group">
        <label for="due" class="form-label">期限</label>
        <input id="due" v-model="form.due" type="datetime-local" class="form-input" required />
      </div>
      <div>
        <button type="button" class="submit-button" @click="submitForm">作成する</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive } from "vue";
import { useRouter } from "vue-router";
const router = useRouter();

const getTomorrow = () => {
  const date = new Date();
  date.setDate(date.getDate() + 1);
  const yyyy = date.getFullYear();
  const mm = String(date.getMonth() + 1).padStart(2, "0");
  const dd = String(date.getDate()).padStart(2, "0");
  const hh = String(date.getHours()).padStart(2, "0");
  const min = String(date.getMinutes()).padStart(2, "0");

  return `${yyyy}-${mm}-${dd}T${hh}:${min}`;
};
const form = reactive({
  name: "",
  choice1: "",
  choice2: "",
  due: getTomorrow(),
});

const submitForm = async () => {
  console.log(form);
  if (!form.name || !form.choice1 || !form.choice2) {
    alert("投票名と選択肢1, 2は必須です。");
    return;
  }
  if (form.choice1 === form.choice2) {
    alert("選択肢1と選択肢2は異なる値である必要があります。");
    return;
  }

  try {
    const res = await fetch("/api/polls", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(form),
    });
    if (res.ok) {
      router.push("/");
      form.name = "";
      form.choice1 = "";
      form.choice2 = "";
      form.due = "";
    } else {
      throw new Error("投票の作成に失敗しました。");
    }
  } catch (error) {
    console.error(error);
    alert("エラーが発生しました。");
  }
};
</script>

<style scoped>
.header {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  margin-bottom: 20px;
}
.back-link {
  align-self: flex-start;
  margin: 10px;
  text-decoration: none;
  color: #ffffff;
}
.page-container {
  text-align: center;
  background-color: #0f172b;
  color: #ffffff;
  min-height: 100vh;
}
.form-container {
  max-width: 400px;
  margin: 0 auto;
  padding: 20px;
  border-radius: 8px;
}
.form-group {
  margin-bottom: 20px;
  display: flex;
  flex-direction: column;
  text-align: left;
}
.form-label {
  margin-bottom: 5px;
  font-weight: bold;
}
.form-input {
  padding: 10px;
  border: 2px solid #90a1b9;
  border-radius: 4px;
  width: 100%;
  background-color: #1d293d;
  color: #ffffff;
}
.form-input::placeholder {
  color: #45556c;
}

.submit-button {
  background-color: #162456;
  color: white;
  padding: 10px 20px;
  border: 1px solid #193cb8;
  border-radius: 4px;
  cursor: pointer;
}
.submit-button:hover {
  background-color: #193cb8;
}
</style>
