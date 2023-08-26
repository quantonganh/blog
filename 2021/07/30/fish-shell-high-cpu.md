---
title: fish shell - very high CPU usage
date: 2021-07-30
description: Today I learned a new command.
categories:
  - Operating System
tags:
  - osx
  - fish
  - sample
---
Sometimes my Macbook is loud like a plow, and the culprit is usually [docker](https://github.com/docker/for-mac/issues?q=high+cpu). This time it's different, `top -u` tells me that [fish shell](https://fishshell.com/) is taking up > 100% CPU.

I just switched from [zsh](https://github.com/zsh-users/zsh) to [fish](https://github.com/fish-shell/fish-shell) about a year ago after having problems with [git completion](https://stackoverflow.com/questions/62766815/git-completion-in-zsh-git-func-wrap3-not-found ).

My favorite feature about fish is probably [autosuggestions](https://fishshell.com/docs/current/tutorial.html#autosuggestions). Normally, when I want to run a command again, I can `control+R`, type a few letters to select the command I want to run. With fish shell, I can type that command and then `->` or `control+F` is fine.

Going back to the problem above, trying to search through Google I found: https://github.com/fish-shell/fish-shell/issues/6353#issuecomment-558360629

I run the command `sample` and the result is:

```shell
Analysis of sampling fish (pid 26885) every 1 millisecond
Process:         fish [26885]
Path:            /usr/local/Cellar/fish/3.2.1/bin/fish
Load Address:    0x107c0a000
Identifier:      fish
Version:         0
Code Type:       X86-64
Parent Process:  ??? [26884]

Date/Time:       2021-07-30 16:06:09.188 +0700
Launch Time:     2021-07-30 15:21:06.317 +0700
OS Version:      Mac OS X 10.15.7 (19H1030)
Report Version:  7
Analysis Tool:   /usr/bin/sample

Physical footprint:         17.6M
Physical footprint (peak):  24.6M
----

Call graph:
    4468 Thread_16282782   DispatchQueue_1: com.apple.main-thread  (serial)
    + 4468 start  (in libdyld.dylib) + 1  [0x7fff6ba11cc9]
    +   4468 main  (in fish) + 5708  [0x107c0d668]
    +     4468 reader_read(parser_t&, int, io_chain_t const&)  (in fish) + 1467  [0x107cd76d0]
    +       4468 parser_t::eval(std::__1::basic_string<wchar_t, std::__1::char_traits<wchar_t>, std::__1::allocator<wchar_t> > const&, io_chain_t const&, std::__1::shared_ptr<job_gr
oup_t> const&, block_type_t)  (in fish) + 126  [0x107cbe684]
    +         4468 parser_t::eval(std::__1::shared_ptr<parsed_source_t const> const&, io_chain_t const&, std::__1::shared_ptr<job_group_t> const&, block_type_t)  (in fish) + 48  [0x
107cbe7f4]
    +           4468 eval_res_t parser_t::eval_node<ast::job_list_t>(std::__1::shared_ptr<parsed_source_t const> const&, ast::job_list_t const&, io_chain_t const&, std::__1::shared_
ptr<job_group_t> const&, block_type_t)  (in fish) + 939  [0x107cbef65]
    +             4468 parse_execution_context_t::eval_node(ast::job_list_t const&, block_t const*)  (in fish) + 203  [0x107cb4c67]
    +               4468 parse_execution_context_t::run_job_list(ast::job_list_t const&, block_t const*)  (in fish) + 59  [0x107caf237]
    +                 4468 parse_execution_context_t::run_job_conjunction(ast::job_conjunction_t const&, block_t const*)  (in fish) + 80  [0x107caf0bc]
    +                   4468 parse_execution_context_t::run_1_job(ast::job_t const&, block_t const*)  (in fish) + 1875  [0x107cb460b]
    +                     4468 exec_job(parser_t&, std::__1::shared_ptr<job_t> const&, io_chain_t const&)  (in fish) + 2383  [0x107c7670a]
    +                       4468 job_t::continue_job(parser_t&, bool)  (in fish) + 986  [0x107cc889e]
    +                         4468 process_mark_finished_children(parser_t&, bool)  (in fish) + 428  [0x107cc6684]
    +                           4468 topic_monitor_t::check(generation_list_t*, bool)  (in fish) + 379  [0x107ce60bd]
    +                             4468 topic_monitor_t::await_gens(generation_list_t const&)  (in fish) + 199  [0x107ce5da1]
    +                               4468 binary_semaphore_t::wait()  (in fish) + 76  [0x107ce55cc]
    +                                 4468 read  (in libsystem_kernel.dylib) + 10  [0x7fff6bb5381e]
    4468 Thread_16286348
    + 4468 thread_start  (in libsystem_pthread.dylib) + 15  [0x7fff6bc11b8b]
    +   4468 _pthread_start  (in libsystem_pthread.dylib) + 148  [0x7fff6bc16109]
    +     4468 thread_pool_t::run_trampoline(void*)  (in fish) + 14  [0x107ca6442]
    +       4468 thread_pool_t::run()  (in fish) + 243  [0x107ca6187]
    +         4468 std::__1::__function::__func<debounce_t::perform_impl(std::__1::function<void ()>, std::__1::function<void ()>)::$_0, std::__1::allocator<debounce_t::perform_impl(std::__1::function<void ()>, std::__1::function<void ()>)::$_0>, void ()>::operator()()  (in fish) + 22  [0x107ca8188]
    +           4468 debounce_t::impl_t::run_next(unsigned long long)  (in fish) + 251  [0x107ca7079]
    +             4468 std::__1::__function::__func<iothread_trampoline_t<std::__1::function<(anonymous namespace)::highlight_result_t ()>, reader_data_t::super_highlight_me_plenty()::$_2, (anonymous namespace)::highlight_result_t>::iothread_trampoline_t(std::__1::function<(anonymous namespace)::highlight_result_t ()> const&, reader_data_t::super_highlight_me_plenty()::$_2 const&)::'lambda'(), std::__1::allocator<iothread_trampoline_t<std::__1::function<(anonymous namespace)::highlight_result_t ()>, reader_data_t::super_highlight_me_plenty()::$_2, (anonymous namespace)::highlight_result_t>::iothread_trampoline_t(std::__1::function<(anonymous namespace)::highlight_result_t ()> const&, reader_data_t::super_highlight_me_plenty()::$_2 const&)::'lambda'()>, void ()>::operator()()  (in fish) + 42  [0x107cdb7c0]
    +               4468 std::__1::__function::__func<get_highlight_performer(parser_t&, std::__1::basic_string<wchar_t, std::__1::char_traits<wchar_t>, std::__1::allocator<wchar_t> > const&, bool)::$_8, std::__1::allocator<get_highlight_performer(parser_t&, std::__1::basic_string<wchar_t, std::__1::char_traits<wchar_t>, std::__1::allocator<wchar_t> > const&, bool)::$_8>, (anonymous namespace)::highlight_result_t ()>::operator()()  (in fish) + 250  [0x107cd9594]
    +                 4468 highlight_shell(std::__1::basic_string<wchar_t, std::__1::char_traits<wchar_t>, std::__1::allocator<wchar_t> > const&, std::__1::vector<highlight_spec_t, std::__1::allocator<highlight_spec_t> >&, operation_context_t const&, bool)  (in fish) + 180  [0x107c898bd]
    +                   4468 highlighter_t::highlight()  (in fish) + 607  [0x107c87f45]
    +                     4468 is_potential_path(std::__1::basic_string<wchar_t, std::__1::char_traits<wchar_t>, std::__1::allocator<wchar_t> > const&, std::__1::vector<std::__1::basic_string<wchar_t, std::__1::char_traits<wchar_t>, std::__1::allocator<wchar_t> >, std::__1::allocator<std::__1::basic_string<wchar_t, std::__1::char_traits<wchar_t>, std::__1::allocator<wchar_t> > > > const&, operation_context_t const&, unsigned int)  (in fish) + 1483  [0x107c8634f]
    +                       4468 wdirname(std::__1::basic_string<wchar_t, std::__1::char_traits<wchar_t>, std::__1::allocator<wchar_t> > const&)  (in fish) + 62  [0x107cee081]
    +                         4468 str2wcstring(char const*)  (in fish) + 21  [0x107c4d709]
    +                           4468 _platform_strlen  (in libsystem_platform.dylib) + 18  [0x7fff6bc07e52]

...

Total number in stack (recursive counted multiple, when >=5):
        9       _platform_strlen  (in libsystem_platform.dylib) + 0  [0x7fff6bc07e40]
        9       _pthread_start  (in libsystem_pthread.dylib) + 148  [0x7fff6bc16109]
        9       debounce_t::impl_t::run_next(unsigned long long)  (in fish) + 251  [0x107ca7079]
        9       highlight_shell(std::__1::basic_string<wchar_t, std::__1::char_traits<wchar_t>, std::__1::allocator<wchar_t> > const&, std::__1::vector<highlight_spec_t, std::__1::allocator<highlight_spec_t> >&, operation_context_t const&, bool)  (in fish) + 180  [0x107c898bd]
        9       highlighter_t::highlight()  (in fish) + 607  [0x107c87f45]
        9       is_potential_path(std::__1::basic_string<wchar_t, std::__1::char_traits<wchar_t>, std::__1::allocator<wchar_t> > const&, std::__1::vector<std::__1::basic_string<wchar_t, std::__1::char_traits<wchar_t>, std::__1::allocator<wchar_t> >, std::__1::allocator<std::__1::basic_string<wchar_t, std::__1::char_traits<wchar_t>, std::__1::allocator<wchar_t> > > > const&, operation_context_t const&, unsigned int)  (in fish) + 1483  [0x107c8634f]
        9       std::__1::__function::__func<debounce_t::perform_impl(std::__1::function<void ()>, std::__1::function<void ()>)::$_0, std::__1::allocator<debounce_t::perform_impl(std::__1::function<void ()>, std::__1::function<void ()>)::$_0>, void ()>::operator()()  (in fish) + 22  [0x107ca8188]
        9       std::__1::__function::__func<get_highlight_performer(parser_t&, std::__1::basic_string<wchar_t, std::__1::char_traits<wchar_t>, std::__1::allocator<wchar_t> > const&, bool)::$_8, std::__1::allocator<get_highlight_performer(parser_t&, std::__1::basic_string<wchar_t, std::__1::char_traits<wchar_t>, std::__1::allocator<wchar_t> > const&, bool)::$_8>, (anonymous namespace)::highlight_result_t ()>::operator()()  (in fish) + 250  [0x107cd9594]
        9       std::__1::__function::__func<iothread_trampoline_t<std::__1::function<(anonymous namespace)::highlight_result_t ()>, reader_data_t::super_highlight_me_plenty()::$_2, (anonymous namespace)::highlight_result_t>::iothread_trampoline_t(std::__1::function<(anonymous namespace)::highlight_result_t ()> const&, reader_data_t::super_highlight_me_plenty()::$_2 const&)::'lambda'(), std::__1::allocator<iothread_trampoline_t<std::__1::function<(anonymous namespace)::highlight_result_t ()>, reader_data_t::super_highlight_me_plenty()::$_2, (anonymous namespace)::highlight_result_t>::iothread_trampoline_t(std::__1::function<(anonymous namespace)::highlight_result_t ()> const&, reader_data_t::super_highlight_me_plenty()::$_2 const&)::'lambda'()>, void ()>::operator()()  (in fish) + 42  [0x107cdb7c0]
        9       str2wcstring(char const*)  (in fish) + 21  [0x107c4d709]
        9       thread_pool_t::run()  (in fish) + 243  [0x107ca6187]
        9       thread_pool_t::run_trampoline(void*)  (in fish) + 14  [0x107ca6442]
        9       thread_start  (in libsystem_pthread.dylib) + 15  [0x7fff6bc11b8b]
        9       wdirname(std::__1::basic_string<wchar_t, std::__1::char_traits<wchar_t>, std::__1::allocator<wchar_t> > const&)  (in fish) + 62  [0x107cee081]
```

*(Honestly, I've been using Mac for almost 8 years and this is the first time I've heard of the `sample` command.)*

I'm not familiar with these stack trace, but it seems to be recursive. Looking more closely on GitHub, I found: https://github.com/fish-shell/fish-shell/issues/7837#issuecomment-803664453

Allow me to quote the explanation of [ridiculousfish](https://github.com/ridiculousfish):

> Ugh, this is bad. There's two issues: dirname is failing and SIGSEGV is being dropped; both are Mac/BSD specific.
> 
> Syntax highlighting tries to determine which arguments may be valid paths, so it can underline them. We pass the string to dirname to get its parent. I believed dirname merely trims the string from the last slash, which is correct on Linux: it cannot error. But on Mac dirname copies the string to internal storage, and will error if the string exceeds PATH_MAX, in which case it returns null.
> 
> So we try to construct a std::string from null and this produces a SIGSEGV. I had thought that if a thread triggered SIGSEGV, the signal would be delivered somewhere (e.g. to the main thread) or else the program would be terminated, and again on Linux this is correct. But on Mac the signal is just dropped; the thread then repeats and gets another SIGSEGV, leading to the infinite loop.
> 
> Fixes need to be:
> 
> _ Don't block the four error signals (SIGBUS, SIGFPE, SIGILL, or SIGSEGV) in background threads. This would allow fish to crash on Mac.
> 
> _ Stop using system provided dirname, it's a big footgun
> 
> _ Optimization: skip path detection for arguments of length > PATH_MAX.