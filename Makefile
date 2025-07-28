commit:
	git commit -m "${m}"

push:
	git push -u origin ${b} 
	
push_all:
	git add . && make commit m="${m}" && make push b="${b}"

push_to_master:
	make push_all m="${m}" b="master"

pm:
	make push_to_master m="${m}"
