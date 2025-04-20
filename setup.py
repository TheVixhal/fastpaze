from setuptools import setup, find_packages

setup(
    name="fastpaze",
    version="0.1.0",
    description="A blazing-fast Python web framework with Go-powered core.",
    author="Vishal Singh Baraiya",
    author_email="realvixhal@gmail.com",
    packages=find_packages(include=["fastpaze", "fastpaze.*"]),
    include_package_data=True,
    package_data={
        "fastpaze": ["*.so", "*.h", "*.mod", "*.sum"],
    },
    zip_safe=False,
    classifiers=[
        "Programming Language :: Python :: 3",
        "Operating System :: OS Independent",
    ],
    install_requires=[],
)
