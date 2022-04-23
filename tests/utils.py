""" Util functions and test helpers """
import json


def create_user_msg(content=""):
    """Create a mock object with all the attributes of the discord obj"""

    # This is a mock class and has no real use anyway.
    # pylint: disable=too-few-public-methods
    class MockChannel:
        """Mock Channel object"""

        def __init__(self):
            # Unfortunately, the class we're mocking used an invalid name :(
            # pylint: disable=invalid-name
            self.id = "baz"

    class MockDiscordObj:
        """Mock Discord object"""

        def __init__(self, content):
            self.author = "foo"
            self.channel = MockChannel()
            self.content = content

    return MockDiscordObj(content)


def create_mock_response(status_code, kind):
    """Create a mock response object depending on the kind"""

    # This is a mock class and has no real use anyway.
    # pylint: disable=too-few-public-methods
    class MockResponse:
        """Mock requests response object"""

        def __init__(self, status_code, kind):
            self.status_code = status_code
            self.content = b"baz"
            if kind == "jeopardy":
                self.text = json.dumps(
                    {
                        "title": "category",
                        "clues": [
                            {
                                "question": "a",
                                "answer": "b",
                            },
                            {
                                "question": "a",
                                "answer": "b",
                            },
                            {
                                "question": "a",
                                "answer": "b",
                            },
                            {
                                "question": "a",
                                "answer": "b",
                            },
                            {
                                "question": "a",
                                "answer": "b",
                            },
                        ],
                    }
                )
            elif kind == "gif":
                self.text = json.dumps(
                    {
                        "data": [
                            {
                                "bitly_gif_url": "foo",
                            },
                            {
                                "bitly_gif_url": "bar",
                            },
                        ]
                    }
                )
            elif kind == "empty_gif":
                self.text = json.dumps({"data": []})
            elif kind == "youtube":
                self.text = json.dumps(
                    {"items": [{"id": {"videoId": "foo"}}, {"id": {"videoId": "bar"}}]}
                )
            elif kind == "empty_youtube":
                self.text = json.dumps({"items": []})

        def json(self):
            """Return mocked response as json"""
            return json.loads(self.text)

    return MockResponse(status_code, kind)
