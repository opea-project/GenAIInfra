# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

import ray

ray.init("ray://localhost:10001")


@ray.remote
def square(x):
    return x * x


numbers = [i for i in range(10)]
results = ray.get([square.remote(x) for x in numbers])

print(results)
