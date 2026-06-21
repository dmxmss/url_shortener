import random
import string
import threading

from locust import HttpUser, task, between

short_urls = []
lock = threading.Lock()


def random_video_id():
    chars = string.ascii_letters + string.digits + "-_"
    return "".join(random.choice(chars) for _ in range(11))


def random_youtube_url():
    return f"https://youtube.com/watch?v={random_video_id()}"


class UrlShortenerUser(HttpUser):

    wait_time = between(0.1, 1)

    @task(1)
    def create_short_url(self):

        payload = {
            "url": random_youtube_url()
        }

        with self.client.post(
            "/api/shorten",
            json=payload,
            catch_response=True
        ) as response:

            if response.status_code != 200 and response.status_code != 201:
                response.failure(
                    f"status={response.status_code}"
                )
                return

            try:
                code = response.json()["short_code"]

                with lock:
                    short_urls.append(code)

            except Exception as e:
                response.failure(str(e))

    @task(9)
    def redirect(self):

        with lock:

            if not short_urls:
                return

            code = random.choice(short_urls)

        self.client.get(
            f"/{code}",
            allow_redirects=False
        )
        print(f"redirect {code}")
